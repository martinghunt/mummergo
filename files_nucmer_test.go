package mummergo

import (
	"os"
	"reflect"
	"testing"
)

func TestReadCoords(t *testing.T) {
	expected := []Alignment{
		MustAlignment(join("61", "900", "1", "840", "840", "840", "99.76", "1000", "840", "1", "1", "test_ref1", "test_qry1", "[CONTAINS]")),
		MustAlignment(join("62", "901", "2", "841", "841", "850", "99.66", "999", "839", "1", "1", "test_ref2", "test_qry2", "[CONTAINS]")),
		MustAlignment(join("63", "902", "3", "842", "842", "860", "99.56", "998", "838", "1", "1", "test_ref3", "test_qry3", "[CONTAINS]")),
	}
	for _, fname := range []string{"testdata/coords_file_test_with_header.coords", "testdata/coords_file_test_no_header.coords"} {
		got, err := ReadCoords(fname)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("%s got %#v want %#v", fname, got, expected)
		}
	}
}

func TestReadSnps(t *testing.T) {
	expected := []Snp{
		MustSnp(join("133", "G", ".", "122", "1", "122", "500", "489", "1", "1", "ref", "qry")),
		MustSnp(join("143", ".", "C", "131", "1", "132", "500", "489", "1", "1", "ref", "qry")),
		MustSnp(join("253", "T", "A", "242", "120", "242", "500", "489", "1", "1", "ref", "qry")),
	}
	for _, fname := range []string{"testdata/snp_file_test_with_header.snps", "testdata/snp_file_test_no_header.snps"} {
		got, err := ReadSnps(fname)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("%s got %#v want %#v", fname, got, expected)
		}
	}
}

func TestRunnerCommandsAndScript(t *testing.T) {
	tests := []struct {
		r    Runner
		want string
	}{
		{NewRunner("ref", "qry", "outfile"), "nucmer -p pre ref qry"},
		{NewRunner("ref", "qry", "outfile", WithBreakLen(42)), "nucmer -p pre -b 42 ref qry"},
		{NewRunner("ref", "qry", "outfile", WithDiagDiff(11)), "nucmer -p pre -D 11 ref qry"},
		{NewRunner("ref", "qry", "outfile", WithDiagDiff(11), WithPromer(true)), "promer -p pre ref qry"},
		{NewRunner("ref", "qry", "outfile", WithMaxMatch(true)), "nucmer -p pre --maxmatch ref qry"},
		{NewRunner("ref", "qry", "outfile", WithMinCluster(42)), "nucmer -p pre -c 42 ref qry"},
		{NewRunner("ref", "qry", "outfile", WithSimplify(false)), "nucmer -p pre --nosimplify ref qry"},
		{NewRunner("ref", "qry", "outfile", WithPromer(true), WithSimplify(false)), "promer -p pre ref qry"},
	}
	for _, tt := range tests {
		if got := tt.r.NucmerCommand("ref", "qry", "pre"); got != tt.want {
			t.Fatalf("NucmerCommand got %q want %q", got, tt.want)
		}
	}

	if got, want := NewRunner("ref", "qry", "outfile", WithMinID(42)).DeltaFilterCommand("infile", "outfile"), "delta-filter -i 42 infile > outfile"; got != want {
		t.Fatalf("DeltaFilterCommand got %q want %q", got, want)
	}
	if got, want := NewRunner("ref", "qry", "outfile", WithCoordsHeader(false)).ShowCoordsCommand("infile", "outfile"), "show-coords -dTlro -H infile > outfile"; got != want {
		t.Fatalf("ShowCoordsCommand got %q want %q", got, want)
	}
	if got, want := NewRunner("ref", "qry", "outfile", WithSnpsHeader(false)).ShowSnpsCommand("infile", "outfile"), "show-snps -TClr -H infile > outfile"; got != want {
		t.Fatalf("ShowSnpsCommand got %q want %q", got, want)
	}
	if got, want := NewRunner("ref", "qry", "outfile", WithShowSnpsC(false)).ShowSnpsCommand("infile", "outfile"), "show-snps -Tlr infile > outfile"; got != want {
		t.Fatalf("ShowSnpsCommand got %q want %q", got, want)
	}

	tmp := t.TempDir() + "/script.sh"
	if err := NewRunner("ref", "qry", "outfile").WriteScript(tmp, "ref", "qry", "outfile"); err != nil {
		t.Fatal(err)
	}
	assertSameFile(t, tmp, "testdata/nucmer_test_write_script_no_snps.sh")
	if err := NewRunner("ref", "qry", "outfile", WithShowSnps(true)).WriteScript(tmp, "ref", "qry", "outfile"); err != nil {
		t.Fatal(err)
	}
	assertSameFile(t, tmp, "testdata/nucmer_test_write_script_with_snps.sh")
}

func assertSameFile(t *testing.T, gotFile, wantFile string) {
	t.Helper()
	got, err := os.ReadFile(gotFile)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(wantFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s != %s\ngot:\n%s\nwant:\n%s", gotFile, wantFile, got, want)
	}
}

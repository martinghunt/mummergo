package mummergo

import (
	"reflect"
	"testing"
)

func TestSnpString(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{join("187", "A", "C", "269", "187", "187", "654", "853", "1", "1", "ref_name", "qry_name"), "187\tA\tC\t269\t654\t853\t1\tref_name\tqry_name"},
		{join("187", "A", "C", "269", "187", "187", "0", "0", "654", "853", "1", "-1", "ref_name", "qry_name"), "187\tA\tC\t269\t654\t853\t-1\tref_name\tqry_name"},
	}
	for _, tt := range tests {
		if got := MustSnp(tt.line).String(); got != tt.want {
			t.Fatalf("Snp.String got %q want %q", got, tt.want)
		}
	}
}

func TestVariantInit(t *testing.T) {
	lines := []string{
		join("42", "T", "A", "42", "42", "42", "1000", "1000", "1", "1", "ref", "ref"),
		join("242", "G", ".", "241", "1", "241", "1000", "1000", "1", "1", "ref", "ref"),
		join("300", ".", "G", "298", "0", "298", "1000", "1000", "1", "1", "ref", "ref"),
	}
	want := []VariantType{SNP, Del, Ins}
	for i, line := range lines {
		if got := MustVariant(MustSnp(line)).Type; got != want[i] {
			t.Fatalf("variant type got %v want %v", got, want[i])
		}
	}
}

func TestVariantTypeString(t *testing.T) {
	tests := []struct {
		in   VariantType
		want string
	}{
		{SNP, "SNP"},
		{Del, "DEL"},
		{Ins, "INS"},
		{VariantType(99), "VariantType(99)"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Fatalf("VariantType.String got %q want %q", got, tt.want)
		}
	}
}

func TestUpdateIndelInsertion(t *testing.T) {
	insertion := MustVariant(MustSnp(join("42", ".", "A", "100", "x", "x", "300", "400", "x", "-1", "ref", "qry")))
	ok, err := insertion.UpdateIndel(MustSnp(join("42", ".", "C", "101", "x", "x", "300", "400", "x", "-1", "ref", "qry")))
	if err != nil || !ok {
		t.Fatalf("UpdateIndel got ok=%v err=%v", ok, err)
	}
	want := Variant{Type: Ins, RefStart: 41, RefEnd: 41, RefLength: 300, RefName: "ref", RefBase: ".", QryStart: 99, QryEnd: 101, QryLength: 400, QryName: "qry", QryBase: "AC", Reverse: true}
	if !reflect.DeepEqual(insertion, want) {
		t.Fatalf("insertion got %#v want %#v", insertion, want)
	}
}

func TestUpdateIndelDeletion(t *testing.T) {
	deletion := MustVariant(MustSnp(join("42", "A", ".", "100", "x", "x", "300", "400", "x", "1", "ref", "qry")))
	ok, err := deletion.UpdateIndel(MustSnp(join("43", "C", ".", "100", "x", "x", "300", "400", "x", "1", "ref", "qry")))
	if err != nil || !ok {
		t.Fatalf("UpdateIndel got ok=%v err=%v", ok, err)
	}
	want := Variant{Type: Del, RefStart: 41, RefEnd: 43, RefLength: 300, RefName: "ref", RefBase: "AC", QryStart: 99, QryEnd: 99, QryLength: 400, QryName: "qry", QryBase: ".", Reverse: false}
	if !reflect.DeepEqual(deletion, want) {
		t.Fatalf("deletion got %#v want %#v", deletion, want)
	}
}

func TestGetAllVariants(t *testing.T) {
	variants, err := GetAllVariants("testdata/snp_file_test_get_all_variants.snps")
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 3 {
		t.Fatalf("got %d variants", len(variants))
	}
	if got, want := variants[0].String(), "125\t127\t500\tref1\tTAC\t124\t124\t497\tqry1\t.\t1"; got != want {
		t.Fatalf("variant[0] got %q want %q", got, want)
	}
	if got, want := variants[1].String(), "386\t386\t500\tref1\tC\t383\t383\t497\tqry1\tT\t1"; got != want {
		t.Fatalf("variant[1] got %q want %q", got, want)
	}
	if got, want := variants[2].String(), "479\t479\t500\tref2\t.\t480\t483\t504\tqry2\tGATA\t1"; got != want {
		t.Fatalf("variant[2] got %q want %q", got, want)
	}
}

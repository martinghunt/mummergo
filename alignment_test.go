package mummergo

import (
	"reflect"
	"strings"
	"testing"
)

func join(fields ...string) string {
	return strings.Join(fields, "\t")
}

func TestAlignmentInitNucmer(t *testing.T) {
	a := MustAlignment(join("1", "100", "2", "200", "101", "202", "42.42", "123", "456", "-1", "0", "ref", "qry", "[FOO]"))
	if a.RefStart != 0 || a.RefEnd != 99 || a.QryStart != 1 || a.QryEnd != 199 ||
		a.HitLengthRef != 101 || a.HitLengthQry != 202 || a.PercentIdentity != 42.42 ||
		a.RefLength != 123 || a.QryLength != 456 || a.Frame != -1 ||
		a.RefName != "ref" || a.QryName != "qry" {
		t.Fatalf("unexpected alignment: %#v", a)
	}
}

func TestAlignmentInitPromer(t *testing.T) {
	a := MustAlignment(join("1", "1398", "4891054", "4892445", "1398", "1392", "89.55", "93.18", "0.21", "1398", "5349013", "1", "1", "ref", "qry", "[CONTAINED]"))
	if a.RefStart != 0 || a.RefEnd != 1397 || a.QryStart != 4891053 || a.QryEnd != 4892444 ||
		a.HitLengthRef != 1398 || a.HitLengthQry != 1392 || a.PercentIdentity != 89.55 ||
		a.RefLength != 1398 || a.QryLength != 5349013 || a.Frame != 1 ||
		a.RefName != "ref" || a.QryName != "qry" {
		t.Fatalf("unexpected alignment: %#v", a)
	}
}

func TestAlignmentSwap(t *testing.T) {
	in := MustAlignment(join("1", "100", "2", "200", "101", "202", "42.42", "123", "456", "-1", "0", "ref", "qry"))
	out := MustAlignment(join("2", "200", "1", "100", "202", "101", "42.42", "456", "123", "-1", "0", "qry", "ref"))
	in.Swap()
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("swap got %#v want %#v", in, out)
	}
	in.Swap()
	if in.String() != "1\t100\t2\t200\t101\t202\t42.42\t123\t456\t-1\tref\tqry" {
		t.Fatalf("swap back got %s", in.String())
	}
}

func TestAlignmentCoordsAndStrand(t *testing.T) {
	for _, h := range []string{
		join("1", "100", "1", "100", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "qry"),
		join("1", "101", "100", "1", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "qry"),
	} {
		if got := MustAlignment(h).QryCoords(); got != (Interval{0, 99}) {
			t.Fatalf("QryCoords got %#v", got)
		}
	}
	for _, h := range []string{
		join("1", "100", "1", "100", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"),
		join("100", "1", "100", "1", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"),
	} {
		if got := MustAlignment(h).RefCoords(); got != (Interval{0, 99}) {
			t.Fatalf("RefCoords got %#v", got)
		}
	}

	tests := []struct {
		line string
		want bool
	}{
		{join("1", "100", "1", "100", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"), true},
		{join("100", "1", "100", "1", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"), true},
		{join("1", "100", "100", "1", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"), false},
		{join("100", "1", "1", "100", "100", "100", "100.00", "1000", "1000", "1", "1", "ref", "ref"), false},
	}
	for _, tt := range tests {
		if got := MustAlignment(tt.line).OnSameStrand(); got != tt.want {
			t.Fatalf("OnSameStrand got %v want %v for %s", got, tt.want, tt.line)
		}
	}
}

func TestAlignmentString(t *testing.T) {
	a := MustAlignment(join("1", "100", "2", "200", "101", "202", "42.42", "123", "456", "-1", "0", "ref", "qry"))
	if got, want := a.String(), "1\t100\t2\t200\t101\t202\t42.42\t123\t456\t-1\tref\tqry"; got != want {
		t.Fatalf("String got %q want %q", got, want)
	}
}

func TestReverseReferenceAndQuery(t *testing.T) {
	aln := MustAlignment(join("100", "142", "1", "42", "43", "42", "100.00", "150", "100", "1", "1", "ref", "qry"))
	aln.ReverseReference()
	if want := MustAlignment(join("51", "9", "1", "42", "43", "42", "100.00", "150", "100", "1", "1", "ref", "qry")); !reflect.DeepEqual(aln, want) {
		t.Fatalf("ReverseReference got %#v want %#v", aln, want)
	}
	aln = MustAlignment(join("100", "142", "1", "42", "43", "42", "100.00", "150", "100", "1", "1", "ref", "qry"))
	aln.ReverseQuery()
	if want := MustAlignment(join("100", "142", "100", "59", "43", "42", "100.00", "150", "100", "1", "1", "ref", "qry")); !reflect.DeepEqual(aln, want) {
		t.Fatalf("ReverseQuery got %#v want %#v", aln, want)
	}
}

func TestIntersectsVariant(t *testing.T) {
	indel := MustVariant(MustSnp("100\tA\t.\t600\t75\t77\t1\t0\t606\t1700\t1\t1\tref\tqry"))
	tests := []struct {
		line string
		want bool
	}{
		{"100\t500\t600\t1000\t501\t501\t100.00\t600\t1700\t1\t1\tref\tqry", true},
		{"101\t500\t600\t1000\t501\t501\t100.00\t600\t1700\t1\t1\tref\tqry", false},
		{"100\t500\t601\t1000\t501\t501\t100.00\t600\t1700\t1\t1\tref\tqry", false},
		{"101\t500\t601\t1000\t501\t501\t100.00\t600\t1700\t1\t1\tref\tqry", false},
	}
	for _, tt := range tests {
		if got := MustAlignment(tt.line).IntersectsVariant(indel); got != tt.want {
			t.Fatalf("IntersectsVariant got %v want %v", got, tt.want)
		}
	}
}

func TestQryCoordsFromRefCoord(t *testing.T) {
	aln := MustAlignment(join("100", "200", "1", "101", "100", "100", "100.00", "300", "300", "1", "1", "ref", "qry"))
	snp0 := MustVariant(MustSnp(join("140", "A", "T", "40", "x", "x", "300", "300", "x", "1", "ref", "qry")))
	del1 := MustVariant(MustSnp(join("140", "A", ".", "40", "x", "x", "300", "300", "x", "1", "ref", "qry")))
	del2 := del1
	if ok, _ := del2.UpdateIndel(MustSnp(join("141", "C", ".", "40", "x", "x", "300", "300", "x", "1", "ref", "qry"))); !ok {
		t.Fatal("expected deletion update")
	}
	ins1 := MustVariant(MustSnp(join("150", ".", "A", "50", "x", "x", "300", "300", "x", "1", "ref", "qry")))
	ins2 := ins1
	for _, s := range []Snp{
		MustSnp(join("150", ".", "C", "51", "x", "x", "300", "300", "x", "1", "ref", "qry")),
		MustSnp(join("150", ".", "G", "52", "x", "x", "300", "300", "x", "1", "ref", "qry")),
	} {
		if ok, _ := ins2.UpdateIndel(s); !ok {
			t.Fatal("expected insertion update")
		}
	}

	tests := []struct {
		refCoord int
		vars     []Variant
		pos      int
		inIndel  bool
	}{
		{99, nil, 0, false},
		{100, nil, 1, false},
		{199, nil, 100, false},
		{119, []Variant{del1}, 20, false},
		{149, []Variant{del2}, 48, false},
		{159, []Variant{ins2}, 63, false},
		{159, []Variant{del2, ins2}, 61, false},
		{139, []Variant{del1}, 39, true},
		{139, []Variant{snp0}, 40, false},
		{149, []Variant{ins1}, 49, true},
	}
	for _, tt := range tests {
		got, inIndel, err := aln.QryCoordsFromRefCoord(tt.refCoord, tt.vars)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.pos || inIndel != tt.inIndel {
			t.Fatalf("QryCoordsFromRefCoord(%d) got (%d,%v) want (%d,%v)", tt.refCoord, got, inIndel, tt.pos, tt.inIndel)
		}
	}
}

func TestRefCoordsFromQryCoord(t *testing.T) {
	aln := MustAlignment("1\t606\t596\t1201\t606\t606\t100.00\t606\t1700\t1\t1\tref\tqry")
	indel := MustVariant(MustSnp("127\tA\t.\t77\t75\t77\t1\t0\t606\t1700\t1\t1\tref\tqry"))
	tests := []struct {
		qryCoord int
		vars     []Variant
		pos      int
	}{
		{595, nil, 0},
		{595, []Variant{indel}, 0},
		{995, nil, 400},
		{995, []Variant{indel}, 400},
		{1200, nil, 605},
		{1200, []Variant{indel}, 605},
	}
	for _, tt := range tests {
		got, inIndel, err := aln.RefCoordsFromQryCoord(tt.qryCoord, tt.vars)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.pos || inIndel {
			t.Fatalf("RefCoordsFromQryCoord(%d) got (%d,%v) want (%d,false)", tt.qryCoord, got, inIndel, tt.pos)
		}
	}
}

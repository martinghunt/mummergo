package mummergo

import (
	"fmt"
	"strconv"
	"strings"
)

type VariantType int

const (
	SNP VariantType = 1
	Del VariantType = 2
	Ins VariantType = 3
)

func (t VariantType) String() string {
	switch t {
	case SNP:
		return "SNP"
	case Del:
		return "DEL"
	case Ins:
		return "INS"
	default:
		return fmt.Sprintf("VariantType(%d)", t)
	}
}

type Variant struct {
	Type      VariantType
	RefStart  int
	RefEnd    int
	RefLength int
	RefName   string
	RefBase   string
	QryStart  int
	QryEnd    int
	QryLength int
	QryName   string
	QryBase   string
	Reverse   bool
}

func NewVariant(s Snp) (Variant, error) {
	v := Variant{
		RefStart:  s.RefPos,
		RefLength: s.RefLength,
		RefName:   s.RefName,
		QryStart:  s.QryPos,
		QryLength: s.QryLength,
		QryName:   s.QryName,
		Reverse:   s.Reverse,
	}

	switch {
	case s.RefBase == "." && s.QryBase != ".":
		v.Type = Ins
		v.RefBase = "."
		v.QryBase = s.QryBase
		v.RefEnd = s.RefPos
		v.QryEnd = s.QryPos + 1
	case s.QryBase == "." && s.RefBase != ".":
		v.Type = Del
		v.RefBase = s.RefBase
		v.QryBase = "."
		v.RefEnd = s.RefPos + 1
		v.QryEnd = s.QryPos
	case s.RefBase != "." && s.QryBase != ".":
		v.Type = SNP
		v.RefBase = s.RefBase
		v.QryBase = s.QryBase
		v.RefEnd = s.RefPos + 1
		v.QryEnd = s.QryPos + 1
	default:
		return Variant{}, fmt.Errorf("error constructing Variant from Snp: %s", s.String())
	}

	return v, nil
}

func (v Variant) refIntervalForIntersection() Interval {
	if v.RefStart == v.RefEnd {
		return Interval{Start: v.RefStart, End: v.RefStart + 1}
	}
	return Interval{Start: v.RefStart, End: v.RefEnd}
}

func (v Variant) qryIntervalForIntersection() Interval {
	if v.QryStart == v.QryEnd {
		return Interval{Start: v.QryStart, End: v.QryStart + 1}
	}
	return Interval{Start: v.QryStart, End: v.QryEnd}
}

func (v Variant) containsRefCoord(pos int) bool {
	if v.RefStart == v.RefEnd {
		return pos == v.RefStart
	}
	return v.RefStart <= pos && pos < v.RefEnd
}

func (v Variant) containsQryCoord(pos int) bool {
	if v.QryStart == v.QryEnd {
		return pos == v.QryStart
	}
	return v.QryStart <= pos && pos < v.QryEnd
}

func MustVariant(s Snp) Variant {
	v, err := NewVariant(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (v Variant) String() string {
	reverse := "1"
	if v.Reverse {
		reverse = "-1"
	}
	return strings.Join([]string{
		strconv.Itoa(v.RefStart + 1),
		strconv.Itoa(variantEndForString(v.RefStart, v.RefEnd)),
		strconv.Itoa(v.RefLength),
		v.RefName,
		v.RefBase,
		strconv.Itoa(v.QryStart + 1),
		strconv.Itoa(variantEndForString(v.QryStart, v.QryEnd)),
		strconv.Itoa(v.QryLength),
		v.QryName,
		v.QryBase,
		reverse,
	}, "\t")
}

func variantEndForString(start, end int) int {
	if start == end {
		return start + 1
	}
	return end
}

func (v *Variant) UpdateIndel(s Snp) (bool, error) {
	newVariant, err := NewVariant(s)
	if err != nil {
		return false, err
	}
	if v.Type != Ins && v.Type != Del ||
		v.Type != newVariant.Type ||
		v.QryName != newVariant.QryName ||
		v.RefName != newVariant.RefName ||
		v.Reverse != newVariant.Reverse {
		return false, nil
	}

	if v.Type == Ins &&
		v.RefStart == newVariant.RefStart &&
		v.QryEnd == newVariant.QryStart {
		v.QryBase += newVariant.QryBase
		v.QryEnd = newVariant.QryEnd
		return true, nil
	}

	if v.Type == Del &&
		v.QryStart == newVariant.QryStart &&
		v.RefEnd == newVariant.RefStart {
		v.RefBase += newVariant.RefBase
		v.RefEnd = newVariant.RefEnd
		return true, nil
	}

	return false, nil
}

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

var VariantTypeNames = map[VariantType]string{
	SNP: "SNP",
	Del: "DEL",
	Ins: "INS",
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
		RefEnd:    s.RefPos,
		RefLength: s.RefLength,
		RefName:   s.RefName,
		QryStart:  s.QryPos,
		QryEnd:    s.QryPos,
		QryLength: s.QryLength,
		QryName:   s.QryName,
		Reverse:   s.Reverse,
	}

	switch {
	case s.RefBase == "." && s.QryBase != ".":
		v.Type = Ins
		v.RefBase = "."
		v.QryBase = s.QryBase
	case s.QryBase == "." && s.RefBase != ".":
		v.Type = Del
		v.RefBase = s.RefBase
		v.QryBase = "."
	case s.RefBase != "." && s.QryBase != ".":
		v.Type = SNP
		v.RefBase = s.RefBase
		v.QryBase = s.QryBase
	default:
		return Variant{}, fmt.Errorf("error constructing Variant from Snp:%s", s.String())
	}

	return v, nil
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
		strconv.Itoa(v.RefEnd + 1),
		strconv.Itoa(v.RefLength),
		v.RefName,
		v.RefBase,
		strconv.Itoa(v.QryStart + 1),
		strconv.Itoa(v.QryEnd + 1),
		strconv.Itoa(v.QryLength),
		v.QryName,
		v.QryBase,
		reverse,
	}, "\t")
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
		v.QryEnd+1 == newVariant.QryStart {
		v.QryBase += newVariant.QryBase
		v.QryEnd++
		return true, nil
	}

	if v.Type == Del &&
		v.QryStart == newVariant.QryStart &&
		v.RefEnd+1 == newVariant.RefStart {
		v.RefBase += newVariant.RefBase
		v.RefEnd++
		return true, nil
	}

	return false, nil
}

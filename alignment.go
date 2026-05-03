package mummergo

import (
	"fmt"
	"strconv"
	"strings"
)

type Alignment struct {
	RefStart        int
	RefEnd          int
	QryStart        int
	QryEnd          int
	HitLengthRef    int
	HitLengthQry    int
	PercentIdentity float64
	RefLength       int
	QryLength       int
	Frame           int
	RefName         string
	QryName         string
}

func NewAlignment(line string) (Alignment, error) {
	fields := strings.Split(strings.TrimRight(line, "\r\n"), "\t")
	parseInt := func(i int) (int, error) {
		if i < 0 || i >= len(fields) {
			return 0, fmt.Errorf("field %d missing", i)
		}
		return strconv.Atoi(fields[i])
	}

	var a Alignment
	var err error
	if a.RefStart, err = parseInt(0); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	if a.RefEnd, err = parseInt(1); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	if a.QryStart, err = parseInt(2); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	if a.QryEnd, err = parseInt(3); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	a.RefStart--
	a.RefEnd--
	a.QryStart--
	a.QryEnd--
	if a.HitLengthRef, err = parseInt(4); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	if a.HitLengthQry, err = parseInt(5); err != nil {
		return Alignment{}, alignmentParseError(line)
	}
	if len(fields) <= 12 {
		return Alignment{}, alignmentParseError(line)
	}
	if a.PercentIdentity, err = strconv.ParseFloat(fields[6], 64); err != nil {
		return Alignment{}, alignmentParseError(line)
	}

	if len(fields) >= 15 {
		if a.RefLength, err = parseInt(9); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		if a.QryLength, err = parseInt(10); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		if a.Frame, err = parseInt(11); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		a.RefName = fields[13]
		a.QryName = fields[14]
	} else {
		if a.RefLength, err = parseInt(7); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		if a.QryLength, err = parseInt(8); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		if a.Frame, err = parseInt(9); err != nil {
			return Alignment{}, alignmentParseError(line)
		}
		a.RefName = fields[11]
		a.QryName = fields[12]
	}

	return a, nil
}

func MustAlignment(line string) Alignment {
	a, err := NewAlignment(line)
	if err != nil {
		panic(err)
	}
	return a
}

func alignmentParseError(line string) error {
	return fmt.Errorf("error reading this nucmer line:\n%s", line)
}

func (a *Alignment) Swap() {
	a.RefStart, a.QryStart = a.QryStart, a.RefStart
	a.RefEnd, a.QryEnd = a.QryEnd, a.RefEnd
	a.HitLengthRef, a.HitLengthQry = a.HitLengthQry, a.HitLengthRef
	a.RefLength, a.QryLength = a.QryLength, a.RefLength
	a.RefName, a.QryName = a.QryName, a.RefName
}

func (a Alignment) QryCoords() Interval {
	return NewInterval(a.QryStart, a.QryEnd)
}

func (a Alignment) RefCoords() Interval {
	return NewInterval(a.RefStart, a.RefEnd)
}

func (a Alignment) OnSameStrand() bool {
	return (a.RefStart < a.RefEnd) == (a.QryStart < a.QryEnd)
}

func (a Alignment) IsSelfHit() bool {
	return a.RefName == a.QryName &&
		a.RefStart == a.QryStart &&
		a.RefEnd == a.QryEnd &&
		a.PercentIdentity == 100
}

func (a *Alignment) ReverseQuery() {
	a.QryStart = a.QryLength - a.QryStart - 1
	a.QryEnd = a.QryLength - a.QryEnd - 1
}

func (a *Alignment) ReverseReference() {
	a.RefStart = a.RefLength - a.RefStart - 1
	a.RefEnd = a.RefLength - a.RefEnd - 1
}

func (a Alignment) String() string {
	return strings.Join([]string{
		strconv.Itoa(a.RefStart + 1),
		strconv.Itoa(a.RefEnd + 1),
		strconv.Itoa(a.QryStart + 1),
		strconv.Itoa(a.QryEnd + 1),
		strconv.Itoa(a.HitLengthRef),
		strconv.Itoa(a.HitLengthQry),
		fmt.Sprintf("%.2f", a.PercentIdentity),
		strconv.Itoa(a.RefLength),
		strconv.Itoa(a.QryLength),
		strconv.Itoa(a.Frame),
		a.RefName,
		a.QryName,
	}, "\t")
}

func (a Alignment) IntersectsVariant(v Variant) bool {
	return NewInterval(v.RefStart, v.RefEnd).Intersects(a.RefCoords()) &&
		NewInterval(v.QryStart, v.QryEnd).Intersects(a.QryCoords())
}

func (a Alignment) QryCoordsFromRefCoord(refCoord int, variants []Variant) (int, bool, error) {
	if a.RefCoords().DistanceToPoint(refCoord) > 0 {
		return 0, false, fmt.Errorf("cannot get query coord in QryCoordsFromRefCoord because given ref_coord %d does not lie in nucmer alignment:\n%s", refCoord, a.String())
	}

	indelIndexes := make([]int, 0)
	for i := range variants {
		if variants[i].Type != Ins && variants[i].Type != Del {
			continue
		}
		if !a.IntersectsVariant(variants[i]) {
			continue
		}
		if variants[i].RefStart <= refCoord && refCoord <= variants[i].RefEnd {
			return variants[i].QryStart, true, nil
		}
		if variants[i].RefStart < refCoord {
			indelIndexes = append(indelIndexes, i)
		}
	}

	distance := refCoord - min(a.RefStart, a.RefEnd)
	for _, i := range indelIndexes {
		if variants[i].Type == Ins {
			distance += len(variants[i].QryBase)
		} else {
			distance -= len(variants[i].RefBase)
		}
	}

	if a.OnSameStrand() {
		return min(a.QryStart, a.QryEnd) + distance, false, nil
	}
	return max(a.QryStart, a.QryEnd) - distance, false, nil
}

func (a Alignment) RefCoordsFromQryCoord(qryCoord int, variants []Variant) (int, bool, error) {
	if a.QryCoords().DistanceToPoint(qryCoord) > 0 {
		return 0, false, fmt.Errorf("cannot get ref coord in RefCoordsFromQryCoord because given qry_coord %d does not lie in nucmer alignment:\n%s", qryCoord, a.String())
	}

	indelIndexes := make([]int, 0)
	for i := range variants {
		if variants[i].Type != Ins && variants[i].Type != Del {
			continue
		}
		if !a.IntersectsVariant(variants[i]) {
			continue
		}
		if variants[i].QryStart <= qryCoord && qryCoord <= variants[i].QryEnd {
			return variants[i].RefStart, true, nil
		}
		if variants[i].QryStart < qryCoord {
			indelIndexes = append(indelIndexes, i)
		}
	}

	distance := qryCoord - min(a.QryStart, a.QryEnd)
	for _, i := range indelIndexes {
		if variants[i].Type == Del {
			distance += len(variants[i].RefBase)
		} else {
			distance -= len(variants[i].QryBase)
		}
	}

	if a.OnSameStrand() {
		return min(a.RefStart, a.RefEnd) + distance, false, nil
	}
	return max(a.RefStart, a.RefEnd) - distance, false, nil
}

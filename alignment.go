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
	parseInt := func(i int, name string) (int, error) {
		if i < 0 || i >= len(fields) {
			return 0, fmt.Errorf("field %q at index %d missing", name, i)
		}
		v, err := strconv.Atoi(fields[i])
		if err != nil {
			return 0, fmt.Errorf("field %q at index %d: %w", name, i, err)
		}
		return v, nil
	}
	parseFloat := func(i int, name string) (float64, error) {
		if i < 0 || i >= len(fields) {
			return 0, fmt.Errorf("field %q at index %d missing", name, i)
		}
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return 0, fmt.Errorf("field %q at index %d: %w", name, i, err)
		}
		return v, nil
	}

	var a Alignment
	var err error
	refStart, err := parseInt(0, "reference start")
	if err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	refEnd, err := parseInt(1, "reference end")
	if err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	qryStart, err := parseInt(2, "query start")
	if err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	qryEnd, err := parseInt(3, "query end")
	if err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	a.RefStart, a.RefEnd = fromMummerOrientedCoords(refStart, refEnd)
	a.QryStart, a.QryEnd = fromMummerOrientedCoords(qryStart, qryEnd)
	if a.HitLengthRef, err = parseInt(4, "reference hit length"); err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	if a.HitLengthQry, err = parseInt(5, "query hit length"); err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}
	if len(fields) <= 12 {
		return Alignment{}, alignmentParseError(line, fmt.Errorf("expected at least 13 tab-delimited fields, got %d", len(fields)))
	}
	if a.PercentIdentity, err = parseFloat(6, "percent identity"); err != nil {
		return Alignment{}, alignmentParseError(line, err)
	}

	if len(fields) >= 15 {
		if a.RefLength, err = parseInt(9, "reference length"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
		}
		if a.QryLength, err = parseInt(10, "query length"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
		}
		if a.Frame, err = parseInt(11, "frame"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
		}
		a.RefName = fields[13]
		a.QryName = fields[14]
	} else {
		if a.RefLength, err = parseInt(7, "reference length"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
		}
		if a.QryLength, err = parseInt(8, "query length"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
		}
		if a.Frame, err = parseInt(9, "frame"); err != nil {
			return Alignment{}, alignmentParseError(line, err)
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

func alignmentParseError(line string, cause error) error {
	return fmt.Errorf("error reading this nucmer line: %w\n%s", cause, line)
}

func fromMummerOrientedCoords(start, end int) (int, int) {
	if start <= end {
		return start - 1, end
	}
	return start - 1, end - 2
}

func toMummerOrientedEnd(start, end int) int {
	if start <= end {
		return end
	}
	return end + 2
}

func (a *Alignment) Swap() {
	a.RefStart, a.QryStart = a.QryStart, a.RefStart
	a.RefEnd, a.QryEnd = a.QryEnd, a.RefEnd
	a.HitLengthRef, a.HitLengthQry = a.HitLengthQry, a.HitLengthRef
	a.RefLength, a.QryLength = a.QryLength, a.RefLength
	a.RefName, a.QryName = a.QryName, a.RefName
}

func (a Alignment) QryCoords() Interval {
	return NewIntervalFromOriented(a.QryStart, a.QryEnd)
}

func (a Alignment) RefCoords() Interval {
	return NewIntervalFromOriented(a.RefStart, a.RefEnd)
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
		strconv.Itoa(toMummerOrientedEnd(a.RefStart, a.RefEnd)),
		strconv.Itoa(a.QryStart + 1),
		strconv.Itoa(toMummerOrientedEnd(a.QryStart, a.QryEnd)),
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
	return v.refIntervalForIntersection().Intersects(a.RefCoords()) &&
		v.qryIntervalForIntersection().Intersects(a.QryCoords())
}

func (a Alignment) QryCoordsFromRefCoord(refCoord int, variants []Variant) (int, bool, error) {
	refRange := a.RefCoords()
	qryRange := a.QryCoords()
	if !refRange.Contains(refCoord) {
		return 0, false, fmt.Errorf("cannot get query coord in QryCoordsFromRefCoord because given ref_coord %d does not lie in nucmer alignment:\n%s", refCoord, a.String())
	}

	indelAdjustment := 0
	for i := range variants {
		if variants[i].Type != Ins && variants[i].Type != Del {
			continue
		}
		if !a.IntersectsVariant(variants[i]) {
			continue
		}
		if variants[i].containsRefCoord(refCoord) {
			return variants[i].QryStart, true, nil
		}
		if variants[i].RefStart < refCoord {
			if variants[i].Type == Ins {
				indelAdjustment += len(variants[i].QryBase)
			} else {
				indelAdjustment -= len(variants[i].RefBase)
			}
		}
	}

	distance := refCoord - refRange.Start + indelAdjustment

	if a.OnSameStrand() {
		return qryRange.Start + distance, false, nil
	}
	return qryRange.End - 1 - distance, false, nil
}

func (a Alignment) RefCoordsFromQryCoord(qryCoord int, variants []Variant) (int, bool, error) {
	qryRange := a.QryCoords()
	refRange := a.RefCoords()
	if !qryRange.Contains(qryCoord) {
		return 0, false, fmt.Errorf("cannot get ref coord in RefCoordsFromQryCoord because given qry_coord %d does not lie in nucmer alignment:\n%s", qryCoord, a.String())
	}

	indelAdjustment := 0
	for i := range variants {
		if variants[i].Type != Ins && variants[i].Type != Del {
			continue
		}
		if !a.IntersectsVariant(variants[i]) {
			continue
		}
		if variants[i].containsQryCoord(qryCoord) {
			return variants[i].RefStart, true, nil
		}
		if variants[i].QryStart < qryCoord {
			if variants[i].Type == Del {
				indelAdjustment += len(variants[i].RefBase)
			} else {
				indelAdjustment -= len(variants[i].QryBase)
			}
		}
	}

	distance := qryCoord - qryRange.Start + indelAdjustment

	if a.OnSameStrand() {
		return refRange.Start + distance, false, nil
	}
	return refRange.End - 1 - distance, false, nil
}

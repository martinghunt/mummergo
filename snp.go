package mummergo

import (
	"fmt"
	"strconv"
	"strings"
)

type Snp struct {
	RefPos    int
	RefBase   string
	QryBase   string
	QryPos    int
	RefLength int
	QryLength int
	Reverse   bool
	RefName   string
	QryName   string
}

func NewSnp(line string) (Snp, error) {
	fields := strings.Split(strings.TrimRight(line, "\r\n"), "\t")
	if len(fields) < 12 {
		return Snp{}, snpParseError(line, fmt.Errorf("expected at least 12 tab-delimited fields, got %d", len(fields)))
	}
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

	refPos, err := parseInt(0, "reference position")
	if err != nil {
		return Snp{}, snpParseError(line, err)
	}
	qryPos, err := parseInt(3, "query position")
	if err != nil {
		return Snp{}, snpParseError(line, err)
	}
	refLengthIndex := len(fields) - 6
	refLength, err := parseInt(refLengthIndex, "reference length")
	if err != nil {
		return Snp{}, snpParseError(line, err)
	}
	qryLengthIndex := len(fields) - 5
	qryLength, err := parseInt(qryLengthIndex, "query length")
	if err != nil {
		return Snp{}, snpParseError(line, err)
	}

	var reverse bool
	reverseIndex := len(fields) - 3
	switch fields[reverseIndex] {
	case "1":
		reverse = false
	case "-1":
		reverse = true
	default:
		return Snp{}, snpParseError(line, fmt.Errorf("field %q at index %d: expected 1 or -1, got %q", "reverse", reverseIndex, fields[reverseIndex]))
	}

	return Snp{
		RefPos:    refPos - 1,
		RefBase:   fields[1],
		QryBase:   fields[2],
		QryPos:    qryPos - 1,
		RefLength: refLength,
		QryLength: qryLength,
		Reverse:   reverse,
		RefName:   fields[len(fields)-2],
		QryName:   fields[len(fields)-1],
	}, nil
}

func MustSnp(line string) Snp {
	s, err := NewSnp(line)
	if err != nil {
		panic(err)
	}
	return s
}

func snpParseError(line string, cause error) error {
	return fmt.Errorf("error constructing Snp from mummer show-snps output at this line: %w\n%s", cause, line)
}

func (s Snp) String() string {
	reverse := "1"
	if s.Reverse {
		reverse = "-1"
	}
	return strings.Join([]string{
		strconv.Itoa(s.RefPos + 1),
		s.RefBase,
		s.QryBase,
		strconv.Itoa(s.QryPos + 1),
		strconv.Itoa(s.RefLength),
		strconv.Itoa(s.QryLength),
		reverse,
		s.RefName,
		s.QryName,
	}, "\t")
}

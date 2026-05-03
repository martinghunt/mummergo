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
		return Snp{}, snpParseError(line)
	}
	refPos, err := strconv.Atoi(fields[0])
	if err != nil {
		return Snp{}, snpParseError(line)
	}
	qryPos, err := strconv.Atoi(fields[3])
	if err != nil {
		return Snp{}, snpParseError(line)
	}
	refLength, err := strconv.Atoi(fields[len(fields)-6])
	if err != nil {
		return Snp{}, snpParseError(line)
	}
	qryLength, err := strconv.Atoi(fields[len(fields)-5])
	if err != nil {
		return Snp{}, snpParseError(line)
	}

	var reverse bool
	switch fields[len(fields)-3] {
	case "1":
		reverse = false
	case "-1":
		reverse = true
	default:
		return Snp{}, snpParseError(line)
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

func snpParseError(line string) error {
	return fmt.Errorf("error constructing Snp from mummer show-snps output at this line:\n%s", line)
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

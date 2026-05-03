package mummergo

import (
	"bufio"
	"os"
	"strings"
)

func ReadCoords(filename string) ([]Alignment, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var alignments []Alignment
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") || !strings.Contains(line, "\t") {
			continue
		}
		a, err := NewAlignment(line)
		if err != nil {
			return nil, err
		}
		alignments = append(alignments, a)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return alignments, nil
}

func ReadSnps(filename string) ([]Snp, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var snps []Snp
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") || !strings.Contains(line, "\t") {
			continue
		}
		s, err := NewSnp(line)
		if err != nil {
			return nil, err
		}
		snps = append(snps, s)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return snps, nil
}

func GetAllVariants(filename string) ([]Variant, error) {
	snps, err := ReadSnps(filename)
	if err != nil {
		return nil, err
	}

	var variants []Variant
	for _, s := range snps {
		if len(variants) > 0 {
			updated, err := variants[len(variants)-1].UpdateIndel(s)
			if err != nil {
				return nil, err
			}
			if updated {
				continue
			}
		}
		v, err := NewVariant(s)
		if err != nil {
			return nil, err
		}
		variants = append(variants, v)
	}
	return variants, nil
}

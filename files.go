package mummergo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const scannerMaxTokenSize = 16 * 1024 * 1024

func ReadCoords(filename string) ([]Alignment, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readCoordsFrom(f, filename)
}

func ReadCoordsFrom(r io.Reader) ([]Alignment, error) {
	return readCoordsFrom(r, "")
}

func readCoordsFrom(r io.Reader, source string) ([]Alignment, error) {
	return readRecords(r, source, NewAlignment)
}

func ReadSnps(filename string) ([]Snp, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readSnpsFrom(f, filename)
}

func ReadSnpsFrom(r io.Reader) ([]Snp, error) {
	return readSnpsFrom(r, "")
}

func readSnpsFrom(r io.Reader, source string) ([]Snp, error) {
	return readRecords(r, source, NewSnp)
}

func GetAllVariants(filename string) ([]Variant, error) {
	snps, err := ReadSnps(filename)
	if err != nil {
		return nil, err
	}

	return variantsFromSnps(snps)
}

func GetAllVariantsFrom(r io.Reader) ([]Variant, error) {
	snps, err := ReadSnpsFrom(r)
	if err != nil {
		return nil, err
	}
	return variantsFromSnps(snps)
}

func variantsFromSnps(snps []Snp) ([]Variant, error) {
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

func readRecords[T any](r io.Reader, source string, parse func(string) (T, error)) ([]T, error) {
	var records []T
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), scannerMaxTokenSize)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if skipMummerHeaderLine(line) {
			continue
		}
		record, err := parse(line)
		if err != nil {
			return nil, sourceLineError(source, lineNo, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, sourceError(source, err)
	}
	return records, nil
}

func skipMummerHeaderLine(line string) bool {
	return strings.HasPrefix(line, "[") || !strings.Contains(line, "\t")
}

func sourceLineError(source string, lineNo int, err error) error {
	if source == "" {
		return fmt.Errorf("line %d: %w", lineNo, err)
	}
	return fmt.Errorf("%s:%d: %w", source, lineNo, err)
}

func sourceError(source string, err error) error {
	if source == "" {
		return err
	}
	return fmt.Errorf("%s: %w", source, err)
}

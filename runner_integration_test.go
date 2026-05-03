//go:build integration

package mummergo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerRun(t *testing.T) {
	for _, name := range []string{"nucmer", "delta-filter", "show-coords", "show-snps"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s not found in PATH", name)
		}
	}

	tmp := t.TempDir() + "/nucmer.out"
	runner := NewRunner(
		"testdata/nucmer_test_ref.fa",
		"testdata/nucmer_test_qry.fa",
		tmp,
		WithCoordsHeader(false),
		WithShowSnps(true),
		WithSnpsHeader(false),
	)
	result, err := runner.RunWithResult()
	if err != nil {
		t.Fatal(err)
	}
	assertSameFile(t, tmp, "testdata/nucmer_test_out.coords")
	assertSameFile(t, tmp+".snps", "testdata/nucmer_test_out.coords.snps")
	if _, err := os.Stat(result.TempDir); !os.IsNotExist(err) {
		t.Fatalf("default RunWithResult should remove temp dir %s, stat err=%v", result.TempDir, err)
	}
}

func TestRunnerRunKeepTemp(t *testing.T) {
	for _, name := range []string{"nucmer", "delta-filter", "show-coords", "show-snps"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s not found in PATH", name)
		}
	}

	parent := t.TempDir()
	outfile := filepath.Join(t.TempDir(), "nucmer.out")
	runner := NewRunner(
		"testdata/nucmer_test_ref.fa",
		"testdata/nucmer_test_qry.fa",
		outfile,
		WithCoordsHeader(false),
		WithShowSnps(true),
		WithSnpsHeader(false),
		WithTempDir(parent),
		WithKeepTemp(true),
	)
	result, err := runner.RunWithResult()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.TempDir, parent+string(os.PathSeparator)) {
		t.Fatalf("temp dir %s is not inside requested parent %s", result.TempDir, parent)
	}
	for _, name := range []string{"p.delta", "p.delta.filter"} {
		if _, err := os.Stat(filepath.Join(result.TempDir, name)); err != nil {
			t.Fatalf("expected kept temp file %s: %v", name, err)
		}
	}
	assertSameFile(t, outfile, "testdata/nucmer_test_out.coords")
	assertSameFile(t, outfile+".snps", "testdata/nucmer_test_out.coords.snps")
}

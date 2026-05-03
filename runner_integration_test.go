//go:build integration

package mummergo

import (
	"os/exec"
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
	if err := runner.Run(); err != nil {
		t.Fatal(err)
	}
	assertSameFile(t, tmp, "testdata/nucmer_test_out.coords")
	assertSameFile(t, tmp+".snps", "testdata/nucmer_test_out.coords.snps")
}

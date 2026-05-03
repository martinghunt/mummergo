package mummergo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Runner struct {
	Ref          string
	Qry          string
	Outfile      string
	MinID        *int
	MinLength    *int
	BreakLen     *int
	CoordsHeader bool
	DiagDiff     *int
	DiagFactor   *int
	MaxGap       *int
	MaxMatch     bool
	MinCluster   *int
	Simplify     bool
	ShowSnps     bool
	SnpsHeader   bool
	Verbose      bool
	Promer       bool
	ShowSnpsC    bool
	TempDir      string
	KeepTemp     bool
}

type RunnerOption func(*Runner)

type RunResult struct {
	TempDir string
}

func NewRunner(ref, qry, outfile string, opts ...RunnerOption) Runner {
	r := Runner{
		Ref:          ref,
		Qry:          qry,
		Outfile:      outfile,
		CoordsHeader: true,
		Simplify:     true,
		SnpsHeader:   true,
		ShowSnpsC:    true,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func WithMinID(v int) RunnerOption         { return func(r *Runner) { r.MinID = &v } }
func WithMinLength(v int) RunnerOption     { return func(r *Runner) { r.MinLength = &v } }
func WithBreakLen(v int) RunnerOption      { return func(r *Runner) { r.BreakLen = &v } }
func WithCoordsHeader(v bool) RunnerOption { return func(r *Runner) { r.CoordsHeader = v } }
func WithDiagDiff(v int) RunnerOption      { return func(r *Runner) { r.DiagDiff = &v } }
func WithDiagFactor(v int) RunnerOption    { return func(r *Runner) { r.DiagFactor = &v } }
func WithMaxGap(v int) RunnerOption        { return func(r *Runner) { r.MaxGap = &v } }
func WithMaxMatch(v bool) RunnerOption     { return func(r *Runner) { r.MaxMatch = v } }
func WithMinCluster(v int) RunnerOption    { return func(r *Runner) { r.MinCluster = &v } }
func WithSimplify(v bool) RunnerOption     { return func(r *Runner) { r.Simplify = v } }
func WithShowSnps(v bool) RunnerOption     { return func(r *Runner) { r.ShowSnps = v } }
func WithSnpsHeader(v bool) RunnerOption   { return func(r *Runner) { r.SnpsHeader = v } }
func WithVerbose(v bool) RunnerOption      { return func(r *Runner) { r.Verbose = v } }
func WithPromer(v bool) RunnerOption       { return func(r *Runner) { r.Promer = v } }
func WithShowSnpsC(v bool) RunnerOption    { return func(r *Runner) { r.ShowSnpsC = v } }
func WithTempDir(v string) RunnerOption    { return func(r *Runner) { r.TempDir = v } }
func WithKeepTemp(v bool) RunnerOption     { return func(r *Runner) { r.KeepTemp = v } }

func (r Runner) NucmerCommand(ref, qry, outprefix string) string {
	name, args := r.nucmerArgs(ref, qry, outprefix)
	return strings.Join(append([]string{name}, args...), " ")
}

func (r Runner) nucmerArgs(ref, qry, outprefix string) (string, []string) {
	command := "nucmer"
	if r.Promer {
		command = "promer"
	}
	args := []string{"-p", outprefix}
	if r.BreakLen != nil {
		args = append(args, "-b", strconv.Itoa(*r.BreakLen))
	}
	if r.DiagDiff != nil && !r.Promer {
		args = append(args, "-D", strconv.Itoa(*r.DiagDiff))
	}
	if r.DiagFactor != nil {
		args = append(args, "-d", strconv.Itoa(*r.DiagFactor))
	}
	if r.MaxGap != nil {
		args = append(args, "-g", strconv.Itoa(*r.MaxGap))
	}
	if r.MaxMatch {
		args = append(args, "--maxmatch")
	}
	if r.MinCluster != nil {
		args = append(args, "-c", strconv.Itoa(*r.MinCluster))
	}
	if !r.Simplify && !r.Promer {
		args = append(args, "--nosimplify")
	}
	args = append(args, ref, qry)
	return command, args
}

func (r Runner) DeltaFilterCommand(infile, outfile string) string {
	args := r.deltaFilterArgs(infile)
	parts := append([]string{"delta-filter"}, args...)
	parts = append(parts, ">", outfile)
	return strings.Join(parts, " ")
}

func (r Runner) deltaFilterArgs(infile string) []string {
	args := []string{}
	if r.MinID != nil {
		args = append(args, "-i", strconv.Itoa(*r.MinID))
	}
	if r.MinLength != nil {
		args = append(args, "-l", strconv.Itoa(*r.MinLength))
	}
	args = append(args, infile)
	return args
}

func (r Runner) ShowCoordsCommand(infile, outfile string) string {
	args := r.showCoordsArgs(infile)
	parts := append([]string{"show-coords"}, args...)
	parts = append(parts, ">", outfile)
	return strings.Join(parts, " ")
}

func (r Runner) showCoordsArgs(infile string) []string {
	args := []string{"-dTlro"}
	if !r.CoordsHeader {
		args = append(args, "-H")
	}
	args = append(args, infile)
	return args
}

func (r Runner) ShowSnpsCommand(infile, outfile string) string {
	args := r.showSnpsArgs(infile)
	parts := append([]string{"show-snps"}, args...)
	parts = append(parts, ">", outfile)
	return strings.Join(parts, " ")
}

func (r Runner) showSnpsArgs(infile string) []string {
	flag := "-Tlr"
	if r.ShowSnpsC {
		flag = "-TClr"
	}
	args := []string{flag}
	if !r.SnpsHeader {
		args = append(args, "-H")
	}
	args = append(args, infile)
	return args
}

func (r Runner) WriteScript(scriptName, ref, qry, outfile string) error {
	lines := []string{
		r.NucmerCommand(ref, qry, "p"),
		r.DeltaFilterCommand("p.delta", "p.delta.filter"),
		r.ShowCoordsCommand("p.delta.filter", outfile),
	}
	if r.ShowSnps {
		lines = append(lines, r.ShowSnpsCommand("p.delta.filter", outfile+".snps"))
	}
	return os.WriteFile(scriptName, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func (r Runner) Run() error {
	_, err := r.RunWithResult()
	return err
}

func (r Runner) RunWithResult() (RunResult, error) {
	ref, err := filepath.Abs(r.Ref)
	if err != nil {
		return RunResult{}, err
	}
	qry, err := filepath.Abs(r.Qry)
	if err != nil {
		return RunResult{}, err
	}
	outfile, err := filepath.Abs(r.Outfile)
	if err != nil {
		return RunResult{}, err
	}

	tmpdir, err := os.MkdirTemp(r.TempDir, "tmp.run_nucmer.")
	if err != nil {
		return RunResult{}, err
	}
	result := RunResult{TempDir: tmpdir}
	if !r.KeepTemp {
		defer os.RemoveAll(tmpdir)
	}

	name, args := r.nucmerArgs(ref, qry, "p")
	if err := r.runCommand(tmpdir, "", name, args...); err != nil {
		return result, err
	}

	filteredDelta := filepath.Join(tmpdir, "p.delta.filter")
	if err := r.runCommand(tmpdir, filteredDelta, "delta-filter", r.deltaFilterArgs("p.delta")...); err != nil {
		return result, err
	}

	if err := r.runCommand(tmpdir, outfile, "show-coords", r.showCoordsArgs("p.delta.filter")...); err != nil {
		return result, err
	}

	if r.ShowSnps {
		if err := r.runCommand(tmpdir, outfile+".snps", "show-snps", r.showSnpsArgs("p.delta.filter")...); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (r Runner) runCommand(dir, stdoutFile, name string, args ...string) error {
	if r.Verbose {
		fmt.Println("Running command:", strings.Join(append([]string{name}, args...), " "))
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	var f *os.File
	var err error
	if stdoutFile == "" {
		cmd.Stdout = &stdout
	} else {
		f, err = os.Create(stdoutFile)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdout = f
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %s\n\nstdout:\n%s\nstderr:\n%s", strings.Join(append([]string{name}, args...), " "), stdout.String(), stderr.String())
	}
	if r.Verbose {
		if stdout.Len() > 0 {
			fmt.Print(stdout.String())
		}
		if stderr.Len() > 0 {
			fmt.Print(stderr.String())
		}
	}
	return nil
}

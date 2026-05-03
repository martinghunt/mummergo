package mummergo

import (
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
}

type RunnerOption func(*Runner)

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

func (r Runner) NucmerCommand(ref, qry, outprefix string) string {
	command := "nucmer"
	if r.Promer {
		command = "promer"
	}
	parts := []string{command, "-p", outprefix}
	if r.BreakLen != nil {
		parts = append(parts, "-b", strconv.Itoa(*r.BreakLen))
	}
	if r.DiagDiff != nil && !r.Promer {
		parts = append(parts, "-D", strconv.Itoa(*r.DiagDiff))
	}
	if r.DiagFactor != nil {
		parts = append(parts, "-d", strconv.Itoa(*r.DiagFactor))
	}
	if r.MaxGap != nil {
		parts = append(parts, "-g", strconv.Itoa(*r.MaxGap))
	}
	if r.MaxMatch {
		parts = append(parts, "--maxmatch")
	}
	if r.MinCluster != nil {
		parts = append(parts, "-c", strconv.Itoa(*r.MinCluster))
	}
	if !r.Simplify && !r.Promer {
		parts = append(parts, "--nosimplify")
	}
	parts = append(parts, ref, qry)
	return strings.Join(parts, " ")
}

func (r Runner) DeltaFilterCommand(infile, outfile string) string {
	parts := []string{"delta-filter"}
	if r.MinID != nil {
		parts = append(parts, "-i", strconv.Itoa(*r.MinID))
	}
	if r.MinLength != nil {
		parts = append(parts, "-l", strconv.Itoa(*r.MinLength))
	}
	parts = append(parts, infile, ">", outfile)
	return strings.Join(parts, " ")
}

func (r Runner) ShowCoordsCommand(infile, outfile string) string {
	parts := []string{"show-coords", "-dTlro"}
	if !r.CoordsHeader {
		parts = append(parts, "-H")
	}
	parts = append(parts, infile, ">", outfile)
	return strings.Join(parts, " ")
}

func (r Runner) ShowSnpsCommand(infile, outfile string) string {
	flag := "-Tlr"
	if r.ShowSnpsC {
		flag = "-TClr"
	}
	parts := []string{"show-snps", flag}
	if !r.SnpsHeader {
		parts = append(parts, "-H")
	}
	parts = append(parts, infile, ">", outfile)
	return strings.Join(parts, " ")
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
	ref, err := filepath.Abs(r.Ref)
	if err != nil {
		return err
	}
	qry, err := filepath.Abs(r.Qry)
	if err != nil {
		return err
	}
	outfile, err := filepath.Abs(r.Outfile)
	if err != nil {
		return err
	}

	tmpdir, err := os.MkdirTemp("", "tmp.run_nucmer.")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	script := filepath.Join(tmpdir, "run_nucmer.sh")
	if err := r.WriteScript(script, ref, qry, outfile); err != nil {
		return err
	}
	cmd := exec.Command("bash", script)
	cmd.Dir = tmpdir
	output, err := cmd.CombinedOutput()
	if r.Verbose {
		fmt.Print(string(output))
	}
	if err != nil {
		return fmt.Errorf("command failed: bash %s\n\noutput:\n%s", script, output)
	}
	return nil
}

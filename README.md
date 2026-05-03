# mummergo

Go library wrapper for [MUMmer](https://mummer.sourceforge.net/), based on the
core behavior of `pymummer`.

The library is intended for Go programs that need to:

- parse `show-coords` output into alignment objects
- parse `show-snps` output into SNP objects
- collapse adjacent insertion/deletion SNP rows into variants
- map reference/query coordinates through an alignment while accounting for indels
- build and run the standard `nucmer`, `delta-filter`, `show-coords`, and
  `show-snps` command pipeline

The parser and string output behavior is tested against fixtures from the
Python wrapper so that generated output matches the original wrapper where the
same functionality exists.

## Install

```sh
go get github.com/martinghunt/mummergo
```

Then import it:

```go
import "github.com/martinghunt/mummergo"
```

## Coordinate Conventions

MUMmer output files use 1-based coordinates. `mummergo` stores coordinates as
0-based integers in structs, matching `pymummer`'s internal behavior.

String methods convert back to the MUMmer-style 1-based output format:

```go
aln, err := mummergo.NewAlignment(line)
if err != nil {
    return err
}

fmt.Println(aln.RefStart) // 0-based internal coordinate
fmt.Println(aln.String()) // 1-based tab-delimited output
```

## Parsing `show-coords`

`NewAlignment` parses one tab-delimited line from `show-coords -dTlro`.
It supports both nucmer and promer style rows.

```go
line := "1\t100\t2\t200\t101\t202\t42.42\t123\t456\t-1\t0\tref\tqry"

aln, err := mummergo.NewAlignment(line)
if err != nil {
    return err
}

fmt.Println(aln.RefName, aln.QryName)
fmt.Println(aln.RefCoords())
fmt.Println(aln.QryCoords())
fmt.Println(aln.OnSameStrand())
```

To read an entire coords file:

```go
alignments, err := mummergo.ReadCoords("out.coords")
if err != nil {
    return err
}

for _, aln := range alignments {
    if aln.IsSelfHit() {
        continue
    }
    fmt.Println(aln.String())
}
```

Header lines and non-tab-delimited lines are skipped, matching the Python
wrapper behavior.

## Parsing `show-snps`

`NewSnp` parses one tab-delimited line from `show-snps`. It supports output
with or without the `-C` option.

```go
line := "187\tA\tC\t269\t187\t187\t654\t853\t1\t1\tref_name\tqry_name"

s, err := mummergo.NewSnp(line)
if err != nil {
    return err
}

fmt.Println(s.RefPos, s.RefBase, s.QryBase, s.QryPos)
fmt.Println(s.String())
```

To read a whole SNP file:

```go
snps, err := mummergo.ReadSnps("out.snps")
if err != nil {
    return err
}

for _, s := range snps {
    fmt.Println(s.String())
}
```

## Variants

MUMmer reports multi-base insertions and deletions as one row per base.
`GetAllVariants` reads a SNP file and collapses adjacent indel rows into single
`Variant` values.

```go
variants, err := mummergo.GetAllVariants("out.snps")
if err != nil {
    return err
}

for _, v := range variants {
    switch v.Type {
    case mummergo.SNP:
        fmt.Println("SNP", v.String())
    case mummergo.Del:
        fmt.Println("deletion", v.RefBase)
    case mummergo.Ins:
        fmt.Println("insertion", v.QryBase)
    }
}
```

You can also create variants manually from parsed SNPs:

```go
s, err := mummergo.NewSnp(line)
if err != nil {
    return err
}

v, err := mummergo.NewVariant(s)
if err != nil {
    return err
}
```

## Coordinate Mapping

`Alignment` can map a coordinate from reference to query, or query to
reference, while accounting for insertion/deletion variants that intersect the
alignment.

```go
variants, err := mummergo.GetAllVariants("out.snps")
if err != nil {
    return err
}

queryPos, inIndel, err := aln.QryCoordsFromRefCoord(refPos, variants)
if err != nil {
    return err
}

if inIndel {
    fmt.Println("reference coordinate lies inside an indel")
}
fmt.Println(queryPos)
```

The reverse direction is:

```go
refPos, inIndel, err := aln.RefCoordsFromQryCoord(queryPos, variants)
```

Both methods return an error if the requested coordinate is outside the
alignment interval.

## Running MUMmer

`Runner` builds and runs this pipeline:

```text
nucmer/promer -> delta-filter -> show-coords -> optional show-snps
```

`Runner.Run` executes each tool directly with `os/exec`. It does not run a
temporary shell script. Intermediate files are written in a temporary directory,
and final coords/SNP outputs are written to the requested output path.

Basic usage:

```go
runner := mummergo.NewRunner(
    "ref.fa",
    "qry.fa",
    "out.coords",
    mummergo.WithCoordsHeader(false),
    mummergo.WithShowSnps(true),
)
err := runner.Run()
if err != nil {
    return err
}
```

The default runner uses:

- `nucmer`, not `promer`
- coords headers enabled
- SNP headers enabled
- `show-snps -TClr`
- simplified nucmer output
- OS default temporary directory
- temporary intermediate files removed after the run

Options are provided with functional options:

```go
runner := mummergo.NewRunner(
    "ref.fa",
    "qry.fa",
    "out.coords",
    mummergo.WithMinID(95),
    mummergo.WithMinLength(100),
    mummergo.WithMaxMatch(true),
    mummergo.WithCoordsHeader(false),
    mummergo.WithShowSnps(true),
    mummergo.WithSnpsHeader(false),
)
```

For large runs, or when the system temp directory is not appropriate, choose
the parent directory used for temporary files:

```go
runner := mummergo.NewRunner(
    "ref.fa",
    "qry.fa",
    "out.coords",
    mummergo.WithTempDir("/scratch/my-run"),
)
```

To keep intermediate files such as `p.delta` and `p.delta.filter`, use
`WithKeepTemp(true)` and call `RunWithResult` to get the generated temp
directory:

```go
runner := mummergo.NewRunner(
    "ref.fa",
    "qry.fa",
    "out.coords",
    mummergo.WithTempDir("/scratch/my-run"),
    mummergo.WithKeepTemp(true),
)

result, err := runner.RunWithResult()
if err != nil {
    return err
}

fmt.Println(result.TempDir)
```

Promer is also supported for command generation/running:

```go
runner := mummergo.NewRunner(
    "ref.fa",
    "qry.fa",
    "out.coords",
    mummergo.WithPromer(true),
)
```

You can inspect the generated commands without running them:

```go
runner := mummergo.NewRunner("ref.fa", "qry.fa", "out.coords")

fmt.Println(runner.NucmerCommand("ref.fa", "qry.fa", "p"))
fmt.Println(runner.DeltaFilterCommand("p.delta", "p.delta.filter"))
fmt.Println(runner.ShowCoordsCommand("p.delta.filter", "out.coords"))
fmt.Println(runner.ShowSnpsCommand("p.delta.filter", "out.coords.snps"))
```

Or write an equivalent shell script for inspection/debugging:

```go
err := runner.WriteScript("run_nucmer.sh", "ref.fa", "qry.fa", "out.coords")
```

## Requirements For `Runner.Run`

`Runner.Run` executes external MUMmer binaries. These must be available on
`PATH`:

- `nucmer` or `promer`
- `delta-filter`
- `show-coords`
- `show-snps` when `WithShowSnps(true)` is used

Parsing functions such as `ReadCoords`, `ReadSnps`, and `GetAllVariants` do not
require MUMmer to be installed.

## Testing

Run:

```sh
go test ./...
```

Tests are translated from the Python wrapper and use the same fixture files in
`testdata/`.

The default test suite does not require MUMmer binaries. To run the end-to-end
test that executes `Runner.Run`, opt into integration tests:

```sh
go test ./... -tags integration
```

The integration test still checks `PATH` first and skips with a clear message if
any required MUMmer binary is missing.

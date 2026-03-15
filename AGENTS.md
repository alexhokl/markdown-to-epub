# AGENTS.md — Coding Agent Instructions

This is a Go CLI tool that converts Markdown files to EPUB format. It uses
[Cobra](https://github.com/spf13/cobra) for CLI structure and
[goldmark](https://github.com/yuin/goldmark) for Markdown parsing.

## Repository Layout

```
main.go          — Entrypoint; calls cmd.Execute()
go.mod / go.sum  — Module definition and checksums
Taskfile.yml     — Task runner (go-task)
cmd/
  root.go        — Root cobra command, Execute(), initConfig()
  generate.go    — "generate" subcommand — all core logic
  style.css      — Embedded CSS (via //go:embed) for EPUB styling
```

## Build, Lint, and Test Commands

The project uses [go-task](https://taskfile.dev) (`task` CLI). All commands
can also be run directly with `go` or `golangci-lint`.

### Build

```bash
task build
# or directly:
go build -o /dev/null ./...
```

### Install

```bash
task install
# or directly:
go install
```

### Test

```bash
task test
# or directly:
go test ./...
```

### Run a Single Test

```bash
# Run one specific test function in a package
go test -v ./cmd/ -run TestFunctionName

# Run all tests matching a pattern
go test -v ./... -run TestPrefix
```

### Test Coverage

```bash
task coverage          # go test --cover ./...
task coverage-html     # generates coverage.html
```

### Benchmarks

```bash
task bench             # go test -bench=. -benchmem ./...
```

### Lint

```bash
task lint
# or directly:
golangci-lint run
```

There is no `.golangci.yml` in the repo; golangci-lint runs with its defaults.

### Security Scan

```bash
task sec
# or directly:
gosec ./...
```

## Code Style Guidelines

### General

- Follow standard idiomatic Go conventions throughout.
- Keep functions small and focused on a single responsibility.
- No panics; propagate errors through return values only.
- No global mutable state beyond package-level cobra command and options vars.

### Imports

Group imports into exactly two blocks separated by a blank line:
1. Standard library packages
2. Third-party packages

```go
import (
    "bytes"
    _ "embed"
    "fmt"
    "os"

    "github.com/alexhokl/helper/cli"
    "github.com/spf13/cobra"
    "github.com/yuin/goldmark"
)
```

Use blank imports (`_ "embed"`) only when required by a compiler directive such
as `//go:embed`. Do not introduce internal packages; all logic lives in `cmd/`.

### Naming Conventions

| Construct | Convention | Example |
|---|---|---|
| Package | single lowercase word | `package cmd` |
| File | lowercase, no underscores | `generate.go` |
| Exported type / func | PascalCase | `Execute()` |
| Unexported func | camelCase | `runGenerate`, `createEpub` |
| Struct | PascalCase | `generateOptions` |
| Struct fields | camelCase, unexported | `markdownFilename`, `epubFilename` |
| Variable | camelCase | `htmlContent`, `cssPath` |
| Cobra command var | camelCase + `Cmd` suffix | `generateCmd`, `rootCmd` |
| Options var | camelCase + `Ops` suffix | `generateOps` |

### Types and Structs

- Define per-command option structs with all unexported fields.
- Declare a package-level variable of the struct type to hold flag values.
- Keep struct definitions immediately before the cobra command variable that
  uses them.

```go
type generateOptions struct {
    markdownFilename string
    epubFilename     string
    overwrite        bool
    title            string
    author           string
    language         string
}

var generateOps generateOptions
```

### Error Handling

- Always use `fmt.Errorf` with `%w` to wrap errors, preserving the chain.
- Error message format: `"verb phrase: %w"` (lowercase, descriptive).
- Use `RunE` (not `Run`) on cobra commands so errors propagate correctly.
- Never use `log.Fatal`, `os.Exit`, or `panic` outside of `main`.
- Always check and handle every error; never use `_` to discard errors from
  functions that return an error unless there is an explicit reason.

```go
content, err := os.ReadFile(generateOps.markdownFilename)
if err != nil {
    return fmt.Errorf("failed to read markdown file: %w", err)
}
```

### Cobra Commands

- Add each subcommand to `rootCmd` in its own `init()` function.
- Mark required flags using `MarkFlagRequired`; log failures with
  `cli.LogUnableToMarkFlagAsRequired` from `github.com/alexhokl/helper/cli`.
- Set `SilenceUsage: true` on the root command to suppress usage on errors.

```go
func init() {
    rootCmd.AddCommand(generateCmd)
    flags := generateCmd.Flags()
    flags.StringVarP(&generateOps.markdownFilename, "input", "i", "", "Path to markdown file")
    if err := generateCmd.MarkFlagRequired("input"); err != nil {
        cli.LogUnableToMarkFlagAsRequired("input", err)
    }
}
```

### Embedded Assets

Use `//go:embed` with a blank import of `"embed"` to embed static files at
compile time. Place the directive immediately before the variable declaration.

```go
//go:embed style.css
var defaultCSS string
```

### Resource Cleanup

Use `defer` for cleanup of temporary resources immediately after creation:

```go
tmpFile, err := os.CreateTemp("", "epub-style-*.css")
if err != nil {
    return fmt.Errorf("failed to create temp CSS file: %w", err)
}
defer os.Remove(tmpFile.Name())
```

### Modern Go APIs

The codebase targets Go 1.25+. Prefer modern stdlib APIs where appropriate:

- `strings.SplitSeq` (Go 1.24+) for iterator-based line splitting
- `strings.CutPrefix` / `strings.CutSuffix` (Go 1.20+) over manual index checks
- `os.CreateTemp` (Go 1.16+) over `ioutil.TempFile`

### Comments

- Use doc comments on all exported functions and types.
- Use inline comments sparingly; prefer self-documenting function and variable
  names.
- Section comments inside functions use a single-line `//` with a capital
  letter and no period: `// Read the Markdown file`.

### Formatting

- All code must be formatted with `gofmt` (enforced by `golangci-lint`).
- Do not manually align struct fields or variable assignments beyond what
  `gofmt` produces.

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/yuin/goldmark` | Markdown parser / renderer |
| `github.com/go-shiori/go-epub` | EPUB creation |
| `github.com/alexhokl/helper` | Shared CLI/IO helpers (`cli`, `iohelper`) |
| `github.com/spf13/viper` | Configuration (indirect via helper) |

When adding new functionality, prefer using these existing dependencies over
introducing new ones. Open a discussion before adding a new direct dependency.

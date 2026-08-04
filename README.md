# Generate Project Skeleton (structgen)

CLI that scans a directory and prints its file tree - as Unicode/ASCII box-drawing, JSON, or injected into a markdown file. Zero dependencies, Go stdlib only.

## Usage

```sh
structgen --path <dir> [flags]
```

| Flag | Description |
|------|-------------|
| `--path <dir>` | Project root to scan. **Required.** |
| `--ascii` | 7-bit ASCII connectors instead of Unicode box-drawing. |
| `--json` | Emit the tree as JSON instead of rendering it. |
| `--hidden` | Include dotfiles/dot-directories. |
| `--exclude <pattern>` | Name or glob to exclude. Repeatable; Comma-separated values. |
| `--include <pattern>` | Name or glob pattern to keep even when excluded or hidden. Repeatable; Comma-separated values. 'include' wins over 'exclude' and the hidden file rule. |
| `--inject <file>` | Write the rendered tree into a markdown file instead of stdout. |

Common junk (`.git`, `node_modules`, `vendor`, `__pycache__`, `*.tmp`, …) is excluded by default; `--include` overrides any exclusion, including the built-ins and the hidden-file rule.

### Examples

```sh
structgen --path .                            # print tree
structgen --path . --hidden --include .env    # dotfiles too
structgen --path . --exclude "*.log,dist"     # extra excludes, glob OK
structgen --path . --json                     # JSON to stdout
structgen --path . --inject README.md         # update the tree below
structgen --path . --ascii                    # ASCII output
```

### Injecting into a README

`--inject` replaces the fenced block between the marker comments below (or appends the block at the end of the file if no markers exist yet). Re-running updates it in place — the rest of the file is untouched. The target file must already exist. Cannot be combined with `--json`.

## Project structure

<!-- structgen:start -->
```
.
├── cmd/
│   ├── main.go
│   └── main_test.go
├── internal/
│   ├── inject/
│   └── tree/
│       ├── filter.go
│       ├── filter_test.go
│       ├── inject.go
│       ├── inject_test.go
│       ├── render.go
│       ├── render_test.go
│       ├── tree.go
│       └── tree_test.go
├── go.mod
└── README.md
```
<!-- structgen:end -->

## Development

```sh
go build ./...   # build
go vet ./...     # static checks
go test ./...    # tests
```

Exit codes: `0` success, `1` runtime error, `2` usage error.
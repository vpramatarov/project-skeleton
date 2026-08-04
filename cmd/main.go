package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vpramatarov/project-skeleton/internal/tree"
)

// exitCode is the process exit status: 0 OK, 1 runtime error, 2 usage error.
type exitCode int

const (
	exitOK exitCode = iota
	exitErr
	exitUsage
)

func exit(c exitCode) {
	os.Exit(int(c))
}

type cliMultiFlag []string

func (m *cliMultiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *cliMultiFlag) Set(v string) error {
	for name := range strings.SplitSeq(v, ",") {
		if name = strings.TrimSpace(name); name != "" {
			*m = append(*m, name)
		}
	}

	return nil
}

func main() {
	commandArgs := os.Args[1:]
	stderr := os.Stderr
	flagSet := flag.NewFlagSet("structgen", flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	flagSet.Usage = func() {
		fmt.Fprintln(stderr, "usage: structgen --path <pathToProject> [--ascii] [--json] [--hidden] [--exclude <pattern>]... [--include <pattern>]... [--inject <file>]")
		flagSet.PrintDefaults()
	}

	projectRoot := flagSet.String("path", "", "Path to project root folder to scan [Required].")
	asJSON := flagSet.Bool("json", false, "output as JSON instead of a printed tree")
	showHidden := flagSet.Bool("hidden", false, "include dotfiles/dot-directories")
	ascii := flagSet.Bool("ascii", false, "render connectors in 7-bit ASCII instead of Unicode box-drawing")
	injectPath := flagSet.String("inject", "", "update the tree section of `file` (i.e. README.md) in place or append to the end if markers are missing, instead of printing to stdout")

	var exclude, include cliMultiFlag
	flagSet.Var(&exclude, "exclude", "name or glob pattern to exclude. Repeatable; Comma-separated values.")
	flagSet.Var(&include, "include", "name or glob pattern to keep even when excluded or hidden. Repeatable; Comma-separated values. 'include' wins over 'exclude' and the hidden file rule.")

	if err := flagSet.Parse(commandArgs); err != nil {
		fmt.Fprintln(stderr, "error: ", err)
		exit(exitUsage)
	}

	if *projectRoot == "" {
		fmt.Fprintln(stderr, "error: Project path must be provided.")
		exit(exitUsage)
	}

	glyphs := tree.Unicode
	if *ascii {
		glyphs = tree.ASCII
	}

	filter := tree.NewFilter(exclude, include, *showHidden)
	fileTree, err := tree.BuildTree(*projectRoot, filter)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		exit(exitErr)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(fileTree); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			exit(exitErr)
		}

		return
	}

	if *injectPath != "" {
		var sb strings.Builder
		sb.WriteString(".\n")
		if err := tree.Render(&sb, fileTree, "", glyphs); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			exit(exitErr)
		}

		if err := tree.InjectInto(*injectPath, sb.String()); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			exit(exitErr)
		}

		fmt.Printf("Injected tree into %s\n", *injectPath)
	}

	fmt.Println(".")
	tree.Render(os.Stdout, fileTree, "", glyphs)
}

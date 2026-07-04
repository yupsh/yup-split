// Command yup-split is the CLI wrapper around github.com/gloo-foo/cmd-split.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-split"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name          = "split"
	flagDelimiter = "delimiter"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `split [OPTIONS] [FILE...]

Read lines and split each line on a delimiter, emitting one field per output
line. Default splits on runs of whitespace. With no FILE, or when FILE is -,
read standard input.`

// spec declares the split wrapper: a file-or-stdin filter with a delimiter flag.
var spec = clix.Spec{
	Name:     name,
	Summary:  "split each input line into multiple output lines on a delimiter",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags returns fresh flag instances. It is a constructor rather than a shared
// slice because urfave/cli flag structs retain per-parse state, so each parse
// (including tests) must build its own.
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{
			Name:    flagDelimiter,
			Aliases: []string{"d"},
			Usage:   "field delimiter for splitting (default: whitespace)",
		},
	}
}

// build maps the invocation to split's pipeline: a file-or-stdin source into the
// split command configured by the flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Split(options(inv.Args)...), nil
}

// options folds the parsed flags into split's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagDelimiter) {
		opts = append(opts, command.SplitDelim(c.String(flagDelimiter)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }

package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-split"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "split"

const flagDelimiter = "delimiter"

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `split [OPTIONS] [FILE...]

Read lines and split each line on a delimiter, emitting one field per output
line. Default splits on runs of whitespace. With no FILE, or when FILE is -,
read standard input.`

// buildVersion is the binary's build version threaded from main's ldflags
// target (`var version`) into the CLI. It is an alias, not a defined type:
// cli.Command.Version is a plain string and must be wired as the bare
// `version` identifier (no conversion) for --version to stay verifiably
// bound to the ldflags symbol.
type buildVersion = string

// run builds and executes the split CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version buildVersion, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newApp(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newApp(version buildVersion, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	// Replace urfave/cli's default --version/-v flag with a --version-only
	// flag, freeing the single-letter -v for command flags while still
	// exposing the injected build version. Done here rather than in func
	// init so construction stays explicit.
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "split each input line into multiple output lines on a delimiter",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flagDelimiter,
				Aliases: []string{"d"},
				Usage:   "field delimiter for splitting (default: whitespace)",
			},
		},
		Action: action(stdin, stdout, fs),
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Split(options(c)...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

func options(c *cli.Command) []any {
	var opts []any
	if c.IsSet(flagDelimiter) {
		opts = append(opts, command.SplitDelim(c.String(flagDelimiter)))
	}
	return opts
}

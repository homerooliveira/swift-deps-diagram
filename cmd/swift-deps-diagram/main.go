package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"swift-deps-diagram/internal/app"
	apperrors "swift-deps-diagram/internal/errors"
)

// runApp allows tests to inject a fake runner.
var runApp = app.Run

type cliOptions struct {
	Path          string
	ProjectPath   string
	WorkspacePath string
	Mode          string
	Format        string
	Output        string
	PNGOutput     string
	IncludeTests  bool
}

func parseFlags(args []string, stderr io.Writer) (cliOptions, error) {
	fs := flag.NewFlagSet("swift-deps-diagram", flag.ContinueOnError)
	fs.SetOutput(stderr)

	opts := cliOptions{}
	fs.StringVar(&opts.Path, "path", ".", "Swift package root containing Package.swift")
	fs.StringVar(&opts.ProjectPath, "project", "", "Optional .xcodeproj path")
	fs.StringVar(&opts.WorkspacePath, "workspace", "", "Optional .xcworkspace path")
	fs.StringVar(&opts.Mode, "mode", "auto", "Input mode: auto|spm|xcode")
	fs.StringVar(&opts.Format, "format", "both", "Output format: mermaid|dot|both")
	fs.StringVar(&opts.Output, "output", "", "Output file path (defaults to stdout)")
	fs.StringVar(&opts.PNGOutput, "png-output", "", "Optional PNG output path rendered using Graphviz dot")
	fs.BoolVar(&opts.IncludeTests, "include-tests", false, "Include test targets in the graph")

	if err := fs.Parse(args); err != nil {
		return cliOptions{}, apperrors.New(apperrors.KindInvalidArgs, "invalid arguments", err)
	}

	switch opts.Format {
	case "mermaid", "dot", "both":
	default:
		return cliOptions{}, apperrors.New(apperrors.KindInvalidArgs, "--format must be one of: mermaid|dot|both", nil)
	}
	switch opts.Mode {
	case "auto", "spm", "xcode":
	default:
		return cliOptions{}, apperrors.New(apperrors.KindInvalidArgs, "--mode must be one of: auto|spm|xcode", nil)
	}
	if opts.ProjectPath != "" && opts.WorkspacePath != "" {
		return cliOptions{}, apperrors.New(apperrors.KindInvalidArgs, "--project and --workspace cannot be used together", nil)
	}

	if fs.NArg() > 0 {
		return cliOptions{}, apperrors.New(apperrors.KindInvalidArgs, "unexpected positional arguments", nil)
	}

	return opts, nil
}

func execute(args []string, stdout, stderr io.Writer) int {
	opts, err := parseFlags(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return apperrors.ExitCode(err)
	}

	runErr := runApp(context.Background(), app.Options{
		PackagePath:   opts.Path,
		ProjectPath:   opts.ProjectPath,
		WorkspacePath: opts.WorkspacePath,
		Mode:          opts.Mode,
		Format:        opts.Format,
		OutputPath:    opts.Output,
		PNGOutput:     opts.PNGOutput,
		IncludeTests:  opts.IncludeTests,
	}, stdout)
	if runErr != nil {
		fmt.Fprintln(stderr, runErr.Error())
		return apperrors.ExitCode(runErr)
	}

	return 0
}

func main() {
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr))
}

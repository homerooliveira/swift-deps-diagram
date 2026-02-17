package app

import (
	"context"
	"io"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/graphviz"
	"swift-deps-diagram/internal/inputresolve"
	"swift-deps-diagram/internal/manifest"
	"swift-deps-diagram/internal/output"
	"swift-deps-diagram/internal/render"
	"swift-deps-diagram/internal/swiftpm"
	"swift-deps-diagram/internal/xcodegraph"
	"swift-deps-diagram/internal/xcodeproj"
)

var dumpPackage = swiftpm.DumpPackage
var decodeManifest = manifest.Decode
var buildGraph = graph.Build
var resolveInput = inputresolve.Resolve
var loadXcodeProject = xcodeproj.Load
var buildXcodeGraph = xcodegraph.Build
var renderMermaid = render.Mermaid
var renderDot = render.Dot
var writeOutput = output.Write
var writePNG = graphviz.WritePNG

// Options configure one CLI execution.
type Options struct {
	PackagePath   string
	ProjectPath   string
	WorkspacePath string
	Mode          string
	Format        string
	OutputPath    string
	PNGOutput     string
	IncludeTests  bool
}

func validateOptions(opts Options) error {
	if opts.PackagePath == "" {
		return apperrors.New(apperrors.KindInvalidArgs, "--path cannot be empty", nil)
	}
	if opts.Mode == "" {
		return apperrors.New(apperrors.KindInvalidArgs, "--mode cannot be empty", nil)
	}
	if !inputresolve.IsValidMode(inputresolve.Mode(opts.Mode)) {
		return apperrors.New(apperrors.KindInvalidArgs, "--mode must be one of: auto|spm|xcode", nil)
	}
	if opts.ProjectPath != "" && opts.WorkspacePath != "" {
		return apperrors.New(apperrors.KindInvalidArgs, "--project and --workspace cannot be used together", nil)
	}
	switch opts.Format {
	case "mermaid", "dot", "both":
	default:
		return apperrors.New(apperrors.KindInvalidArgs, "--format must be one of: mermaid|dot|both", nil)
	}
	return nil
}

func renderOutput(g graph.Graph, format string) (string, error) {
	switch format {
	case "mermaid":
		return renderMermaid(g)
	case "dot":
		return renderDot(g)
	case "both":
		mermaidOut, err := renderMermaid(g)
		if err != nil {
			return "", err
		}
		dotOut, err := renderDot(g)
		if err != nil {
			return "", err
		}
		return strings.Join([]string{mermaidOut, dotOut}, "\n\n---\n\n"), nil
	default:
		return "", apperrors.New(apperrors.KindInvalidArgs, "unsupported format", nil)
	}
}

// Run executes the full workflow from manifest dump to emitted diagram output.
func Run(ctx context.Context, opts Options, stdout io.Writer) error {
	if err := validateOptions(opts); err != nil {
		return err
	}
	resolved, err := resolveInput(inputresolve.Request{
		Path:          opts.PackagePath,
		Mode:          inputresolve.Mode(opts.Mode),
		ProjectPath:   opts.ProjectPath,
		WorkspacePath: opts.WorkspacePath,
	})
	if err != nil {
		return err
	}

	var g graph.Graph
	switch resolved.Mode {
	case inputresolve.ModeSPM:
		manifestJSON, err := dumpPackage(ctx, resolved.PackagePath)
		if err != nil {
			return err
		}

		pkg, err := decodeManifest(manifestJSON)
		if err != nil {
			return err
		}

		g, err = buildGraph(pkg, opts.IncludeTests)
		if err != nil {
			return apperrors.New(apperrors.KindRuntime, "failed to build dependency graph", err)
		}
	case inputresolve.ModeXcode:
		project, err := loadXcodeProject(ctx, resolved.ProjectPath)
		if err != nil {
			return err
		}
		g, err = buildXcodeGraph(project, opts.IncludeTests)
		if err != nil {
			return apperrors.New(apperrors.KindRuntime, "failed to build xcode dependency graph", err)
		}
	default:
		return apperrors.New(apperrors.KindInvalidArgs, "unsupported resolved input mode", nil)
	}

	rendered, err := renderOutput(g, opts.Format)
	if err != nil {
		return err
	}

	if err := writeOutput(rendered, opts.OutputPath, stdout); err != nil {
		return err
	}
	if opts.PNGOutput != "" {
		dotOut, err := renderDot(g)
		if err != nil {
			return err
		}
		if err := writePNG(ctx, dotOut, opts.PNGOutput); err != nil {
			return err
		}
	}

	return nil
}

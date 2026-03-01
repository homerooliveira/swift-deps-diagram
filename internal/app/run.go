package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"swift-deps-diagram/internal/bazel"
	"swift-deps-diagram/internal/bazelgraph"
	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/graphviz"
	"swift-deps-diagram/internal/inputresolve"
	"swift-deps-diagram/internal/manifest"
	"swift-deps-diagram/internal/output"
	"swift-deps-diagram/internal/render"
	"swift-deps-diagram/internal/swiftpm"
	"swift-deps-diagram/internal/tuist"
	"swift-deps-diagram/internal/xcodegraph"
	"swift-deps-diagram/internal/xcodeproj"
)

var dumpPackage = swiftpm.DumpPackage
var decodeManifest = manifest.Decode
var buildGraph = graph.Build
var resolveInput = inputresolve.Resolve
var loadXcodeProject = xcodeproj.Load
var generateTuistProject = tuist.Generate
var buildXcodeGraph = xcodegraph.Build
var loadBazelWorkspace = bazel.LoadWorkspace
var buildBazelGraph = bazelgraph.Build
var renderMermaid = render.Mermaid
var renderDot = render.Dot
var renderTerminal = render.Terminal
var writeOutput = output.Write
var writePNG = graphviz.WritePNG
var logInfof = func(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// Options configure one CLI execution.
type Options struct {
	PackagePath   string
	ProjectPath   string
	WorkspacePath string
	BazelTargets  string
	Mode          string
	Format        string
	OutputPath    string
	Verbose       bool
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
		return apperrors.New(apperrors.KindInvalidArgs, "--mode must be one of: auto|spm|xcode|bazel", nil)
	}
	if opts.ProjectPath != "" && opts.WorkspacePath != "" {
		return apperrors.New(apperrors.KindInvalidArgs, "--project and --workspace cannot be used together", nil)
	}
	switch opts.Format {
	case "mermaid", "dot", "png", "terminal":
	default:
		return apperrors.New(apperrors.KindInvalidArgs, "--format must be one of: mermaid|dot|png|terminal", nil)
	}
	return nil
}

func renderTextOutput(g graph.Graph, format string) (string, error) {
	switch format {
	case "mermaid":
		return renderMermaid(g)
	case "dot":
		return renderDot(g)
	case "terminal":
		return renderTerminal(g)
	default:
		return "", apperrors.New(apperrors.KindInvalidArgs, "unsupported format", nil)
	}
}

func absolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	return absPath
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
		BazelTargets:  opts.BazelTargets,
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
		if resolved.TuistPath != "" {
			if err := generateTuistProject(ctx, resolved.TuistPath); err != nil {
				return err
			}
			generated, err := resolveInput(inputresolve.Request{Path: resolved.TuistPath, Mode: inputresolve.ModeXcode})
			if err != nil {
				return err
			}
			if generated.ProjectPath == "" {
				return apperrors.New(apperrors.KindRuntime, fmt.Sprintf("tuist generation completed but no xcode project was resolved at %s", resolved.TuistPath), nil)
			}
			resolved = generated
		}
		project, err := loadXcodeProject(ctx, resolved.ProjectPath)
		if err != nil {
			return err
		}
		g, err = buildXcodeGraph(project, opts.IncludeTests)
		if err != nil {
			return apperrors.New(apperrors.KindRuntime, "failed to build xcode dependency graph", err)
		}
	case inputresolve.ModeBazel:
		workspace, err := loadBazelWorkspace(ctx, resolved.BazelWorkspacePath, resolved.BazelTargets)
		if err != nil {
			return err
		}
		g, err = buildBazelGraph(workspace, opts.IncludeTests)
		if err != nil {
			return apperrors.New(apperrors.KindRuntime, "failed to build bazel dependency graph", err)
		}
	default:
		return apperrors.New(apperrors.KindInvalidArgs, "unsupported resolved input mode", nil)
	}

	if opts.Format == "png" {
		dotOut, err := renderDot(g)
		if err != nil {
			return err
		}
		pngOutputPath := opts.OutputPath
		if pngOutputPath == "" {
			pngOutputPath = "deps.png"
		}
		if err := writePNG(ctx, dotOut, pngOutputPath); err != nil {
			return err
		}
		logInfof("generated png using dot format at %s", absolutePath(pngOutputPath))
		return nil
	}

	rendered, err := renderTextOutput(g, opts.Format)
	if err != nil {
		return err
	}

	if err := writeOutput(rendered, opts.OutputPath, stdout); err != nil {
		return err
	}
	if opts.Verbose && opts.OutputPath != "" {
		switch opts.Format {
		case "mermaid":
			logInfof("generated mermaid content at %s", opts.OutputPath)
		case "dot":
			logInfof("generated dot content at %s", opts.OutputPath)
		case "terminal":
			logInfof("generated terminal content at %s", opts.OutputPath)
		}
	}

	return nil
}

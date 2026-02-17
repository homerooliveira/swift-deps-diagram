package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
	"swift-deps-diagram/internal/graph"
	"swift-deps-diagram/internal/manifest"
	"swift-deps-diagram/internal/output"
	"swift-deps-diagram/internal/render"
	"swift-deps-diagram/internal/swiftpm"
)

var dumpPackage = swiftpm.DumpPackage
var decodeManifest = manifest.Decode
var buildGraph = graph.Build
var renderMermaid = render.Mermaid
var renderDot = render.Dot
var writeOutput = output.Write

// Options configure one CLI execution.
type Options struct {
	PackagePath  string
	Format       string
	OutputPath   string
	IncludeTests bool
}

func validateOptions(opts Options) error {
	if opts.PackagePath == "" {
		return apperrors.New(apperrors.KindInvalidArgs, "--path cannot be empty", nil)
	}
	switch opts.Format {
	case "mermaid", "dot", "both":
	default:
		return apperrors.New(apperrors.KindInvalidArgs, "--format must be one of: mermaid|dot|both", nil)
	}
	return nil
}

func ensureManifestExists(packagePath string) error {
	manifestPath := filepath.Join(packagePath, "Package.swift")
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			return apperrors.New(apperrors.KindManifestNotFound, fmt.Sprintf("Package.swift not found at %s", manifestPath), err)
		}
		return apperrors.New(apperrors.KindManifestNotFound, fmt.Sprintf("unable to access %s", manifestPath), err)
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
	if err := ensureManifestExists(opts.PackagePath); err != nil {
		return err
	}

	manifestJSON, err := dumpPackage(ctx, opts.PackagePath)
	if err != nil {
		return err
	}

	pkg, err := decodeManifest(manifestJSON)
	if err != nil {
		return err
	}

	g, err := buildGraph(pkg, opts.IncludeTests)
	if err != nil {
		return apperrors.New(apperrors.KindRuntime, "failed to build dependency graph", err)
	}

	rendered, err := renderOutput(g, opts.Format)
	if err != nil {
		return err
	}

	if err := writeOutput(rendered, opts.OutputPath, stdout); err != nil {
		return err
	}

	return nil
}

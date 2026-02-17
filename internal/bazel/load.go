package bazel

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

const queryTimeout = 2 * time.Minute

var lookPath = exec.LookPath

var runCommand = func(ctx context.Context, dir string, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func LoadWorkspace(ctx context.Context, workspacePath, scope string) (Workspace, error) {
	if workspacePath == "" {
		return Workspace{}, apperrors.New(apperrors.KindBazelWorkspaceNotFound, "bazel workspace path cannot be empty", nil)
	}
	scope = normalizeScope(scope)

	binary, err := resolveBazelBinary()
	if err != nil {
		return Workspace{}, err
	}

	ruleExpr := fmt.Sprintf(`kind("rule", %s)`, scope)
	labelsOut, err := runQuery(ctx, workspacePath, binary, ruleExpr, "label")
	if err != nil {
		return Workspace{}, err
	}
	labels := parseLabelLines(labelsOut)

	kindOut, err := runQuery(ctx, workspacePath, binary, ruleExpr, "label_kind")
	if err != nil {
		return Workspace{}, err
	}
	kindByLabel, err := parseLabelKinds(kindOut)
	if err != nil {
		return Workspace{}, apperrors.New(apperrors.KindBazelParseFailed, "failed to parse bazel label_kind output", err)
	}

	targets := make([]Target, 0, len(labels))
	for _, label := range labels {
		depsExpr := fmt.Sprintf(`kind("rule", deps(%s, 1))`, label)
		depsOut, err := runQuery(ctx, workspacePath, binary, depsExpr, "label")
		if err != nil {
			return Workspace{}, err
		}

		kind := kindByLabel[label]
		if kind == "" {
			kind = "rule"
		}

		deps := parseLabelLines(depsOut)
		filteredDeps := make([]string, 0, len(deps))
		for _, dep := range deps {
			if dep == label {
				continue
			}
			if !strings.HasPrefix(dep, "//") && !strings.HasPrefix(dep, "@") {
				continue
			}
			filteredDeps = append(filteredDeps, dep)
		}
		filteredDeps = uniqueSorted(filteredDeps)

		targets = append(targets, Target{
			Label: label,
			Kind:  kind,
			Deps:  filteredDeps,
		})
	}

	return Workspace{
		Path:    workspacePath,
		Scope:   scope,
		Targets: targets,
	}, nil
}

func resolveBazelBinary() (string, error) {
	if _, err := lookPath("bazel"); err == nil {
		return "bazel", nil
	}
	if _, err := lookPath("bazelisk"); err == nil {
		return "bazelisk", nil
	}
	return "", apperrors.New(apperrors.KindBazelBinaryNotFound, "neither bazel nor bazelisk was found in PATH", nil)
}

func normalizeScope(scope string) string {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return "//..."
	}
	return scope
}

func runQuery(ctx context.Context, workspacePath, binary, expr, output string) ([]byte, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	args := []string{"query", expr}
	if output != "" {
		args = append(args, "--output="+output)
	}
	args = append(args, "--noimplicit_deps", "--notool_deps")

	stdout, stderr, err := runCommand(queryCtx, workspacePath, binary, args...)
	if err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, apperrors.New(apperrors.KindBazelQueryFailed, "bazel query timed out", queryCtx.Err())
		}
		detail := strings.TrimSpace(string(stderr))
		if detail == "" {
			detail = err.Error()
		}
		return nil, apperrors.New(apperrors.KindBazelQueryFailed, fmt.Sprintf("bazel query failed: %s", detail), err)
	}

	return stdout, nil
}

func parseLabelLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return uniqueSorted(out)
}

func parseLabelKinds(data []byte) (map[string]string, error) {
	lines := strings.Split(string(data), "\n")
	out := make(map[string]string, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " rule ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label_kind line %q", line)
		}
		kind := strings.TrimSpace(parts[0])
		label := strings.TrimSpace(parts[1])
		if kind == "" || label == "" {
			return nil, fmt.Errorf("invalid label_kind line %q", line)
		}
		out[label] = kind
	}
	return out, nil
}

func uniqueSorted(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, value := range in {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

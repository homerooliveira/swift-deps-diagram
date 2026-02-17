package bazel

import (
	"context"
	"errors"
	"strings"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
)

func parseQueryArgs(args []string) (expr string, output string, flags map[string]bool) {
	flags = map[string]bool{
		"--noimplicit_deps": false,
		"--notool_deps":     false,
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "--output=") {
			output = strings.TrimPrefix(arg, "--output=")
			continue
		}
		if _, ok := flags[arg]; ok {
			flags[arg] = true
		}
	}
	if len(args) > 1 {
		expr = args[1]
	}
	return expr, output, flags
}

func TestLoadWorkspaceSuccess(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	})

	lookPath = func(name string) (string, error) {
		if name == "bazel" {
			return "/usr/bin/bazel", nil
		}
		return "", errors.New("not found")
	}

	runCommand = func(_ context.Context, _ string, _ string, args ...string) ([]byte, []byte, error) {
		expr, output, flags := parseQueryArgs(args)
		if !flags["--noimplicit_deps"] || !flags["--notool_deps"] {
			return nil, []byte("missing query filter flags"), errors.New("exit 1")
		}
		switch {
		case expr == `kind("rule", //app:all)` && output == "label":
			return []byte("//app:app\n//app:feature\n//app:feature_test\n"), nil, nil
		case expr == `kind("rule", //app:all)` && output == "label_kind":
			return []byte("swift_binary rule //app:app\nswift_library rule //app:feature\nswift_test rule //app:feature_test\n"), nil, nil
		case expr == `kind("rule", deps(//app:app, 1))` && output == "label":
			return []byte("//app:app\n//app:feature\n@repo//libs:external\n"), nil, nil
		case expr == `kind("rule", deps(//app:feature, 1))` && output == "label":
			return []byte("//app:feature\n"), nil, nil
		case expr == `kind("rule", deps(//app:feature_test, 1))` && output == "label":
			return []byte("//app:feature_test\n//app:feature\n"), nil, nil
		default:
			return nil, []byte("unexpected query"), errors.New("exit 1")
		}
	}

	ws, err := LoadWorkspace(context.Background(), "/tmp/ws", "//app:all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Scope != "//app:all" {
		t.Fatalf("unexpected scope %q", ws.Scope)
	}
	if len(ws.Targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(ws.Targets))
	}

	if ws.Targets[0].Label != "//app:app" {
		t.Fatalf("unexpected target order: %#v", ws.Targets)
	}
	if got := ws.Targets[0].Deps; len(got) != 2 || got[0] != "//app:feature" || got[1] != "@repo//libs:external" {
		t.Fatalf("unexpected deps for //app:app: %#v", got)
	}
}

func TestLoadWorkspaceMissingBinary(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	})

	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, nil, nil
	}

	_, err := LoadWorkspace(context.Background(), "/tmp/ws", "//...")
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindBazelBinaryNotFound) {
		t.Fatalf("expected KindBazelBinaryNotFound, got %v", err)
	}
}

func TestLoadWorkspaceQueryFailureIncludesStderr(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	})

	lookPath = func(name string) (string, error) {
		if name == "bazel" {
			return "/usr/bin/bazel", nil
		}
		return "", errors.New("not found")
	}
	runCommand = func(_ context.Context, _ string, _ string, _ ...string) ([]byte, []byte, error) {
		return nil, []byte("query syntax error"), errors.New("exit 1")
	}

	_, err := LoadWorkspace(context.Background(), "/tmp/ws", "//...")
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindBazelQueryFailed) {
		t.Fatalf("expected KindBazelQueryFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "query syntax error") {
		t.Fatalf("expected stderr in message, got %q", err.Error())
	}
}

func TestLoadWorkspaceDefaultsScope(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	})

	lookPath = func(name string) (string, error) {
		if name == "bazel" {
			return "/usr/bin/bazel", nil
		}
		return "", errors.New("not found")
	}

	firstExpr := ""
	runCommand = func(_ context.Context, _ string, _ string, args ...string) ([]byte, []byte, error) {
		expr, output, _ := parseQueryArgs(args)
		if firstExpr == "" {
			firstExpr = expr
		}
		switch {
		case expr == `kind("rule", //...)` && output == "label":
			return []byte("//app:lib\n"), nil, nil
		case expr == `kind("rule", //...)` && output == "label_kind":
			return []byte("swift_library rule //app:lib\n"), nil, nil
		case expr == `kind("rule", deps(//app:lib, 1))` && output == "label":
			return []byte("//app:lib\n"), nil, nil
		default:
			return nil, []byte("unexpected query"), errors.New("exit 1")
		}
	}

	ws, err := LoadWorkspace(context.Background(), "/tmp/ws", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Scope != "//..." {
		t.Fatalf("expected //... scope, got %q", ws.Scope)
	}
	if firstExpr != `kind("rule", //...)` {
		t.Fatalf("expected default scope query, got %q", firstExpr)
	}
}

func TestLoadWorkspaceLabelKindParseFailure(t *testing.T) {
	oldLookPath := lookPath
	oldRunCommand := runCommand
	t.Cleanup(func() {
		lookPath = oldLookPath
		runCommand = oldRunCommand
	})

	lookPath = func(name string) (string, error) {
		if name == "bazel" {
			return "/usr/bin/bazel", nil
		}
		return "", errors.New("not found")
	}

	runCommand = func(_ context.Context, _ string, _ string, args ...string) ([]byte, []byte, error) {
		expr, output, _ := parseQueryArgs(args)
		switch {
		case expr == `kind("rule", //...)` && output == "label":
			return []byte("//app:lib\n"), nil, nil
		case expr == `kind("rule", //...)` && output == "label_kind":
			return []byte("invalid output line"), nil, nil
		default:
			return nil, nil, nil
		}
	}

	_, err := LoadWorkspace(context.Background(), "/tmp/ws", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !apperrors.IsKind(err, apperrors.KindBazelParseFailed) {
		t.Fatalf("expected KindBazelParseFailed, got %v", err)
	}
}

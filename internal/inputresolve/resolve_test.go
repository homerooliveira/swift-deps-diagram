package inputresolve

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestResolveSPMMode(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// test"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	resolved, err := Resolve(Request{Path: dir, Mode: ModeSPM})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeSPM {
		t.Fatalf("expected mode spm, got %s", resolved.Mode)
	}
	if resolved.PackagePath != dir {
		t.Fatalf("unexpected package path %s", resolved.PackagePath)
	}
}

func TestResolveAutoPrefersXcodeOverSPM(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// test"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}
	projDir := filepath.Join(dir, "Sample.xcodeproj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	resolved, err := Resolve(Request{Path: dir, Mode: ModeAuto})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeXcode {
		t.Fatalf("expected mode xcode, got %s", resolved.Mode)
	}
	if resolved.ProjectPath != projDir {
		t.Fatalf("unexpected project path %s", resolved.ProjectPath)
	}
}

func TestResolveAutoPrefersXcodeOverBazelAndSPM(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// test"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "WORKSPACE.bazel"), []byte("workspace(name = \"x\")"), 0o644); err != nil {
		t.Fatalf("failed to create workspace marker: %v", err)
	}
	projDir := filepath.Join(dir, "Sample.xcodeproj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	resolved, err := Resolve(Request{Path: dir, Mode: ModeAuto})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeXcode {
		t.Fatalf("expected mode xcode, got %s", resolved.Mode)
	}
}

func TestResolveAutoChoosesBazelWhenNoXcode(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// test"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "MODULE.bazel"), []byte("module(name = \"x\")"), 0o644); err != nil {
		t.Fatalf("failed to create module marker: %v", err)
	}

	resolved, err := Resolve(Request{Path: dir, Mode: ModeAuto})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeBazel {
		t.Fatalf("expected mode bazel, got %s", resolved.Mode)
	}
	if resolved.BazelWorkspacePath != dir {
		t.Fatalf("unexpected workspace path %s", resolved.BazelWorkspacePath)
	}
	if resolved.BazelTargets != "//..." {
		t.Fatalf("expected default bazel targets //..., got %s", resolved.BazelTargets)
	}
}

func TestResolveModeValidation(t *testing.T) {
	_, err := Resolve(Request{Path: t.TempDir(), Mode: Mode("bad")})
	if err == nil {
		t.Fatal("expected invalid mode error")
	}
	if !apperrors.IsKind(err, apperrors.KindInvalidArgs) {
		t.Fatalf("expected invalid args kind, got %v", err)
	}
}

func TestResolveBazelModeRequiresWorkspaceMarkers(t *testing.T) {
	dir := t.TempDir()

	_, err := Resolve(Request{Path: dir, Mode: ModeBazel})
	if err == nil {
		t.Fatal("expected bazel workspace not found error")
	}
	if !apperrors.IsKind(err, apperrors.KindBazelWorkspaceNotFound) {
		t.Fatalf("expected bazel workspace not found kind, got %v", err)
	}
}

func TestResolveBazelModeWithTargets(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "WORKSPACE"), []byte("workspace(name = \"x\")"), 0o644); err != nil {
		t.Fatalf("failed to create workspace marker: %v", err)
	}

	resolved, err := Resolve(Request{
		Path:         dir,
		Mode:         ModeBazel,
		BazelTargets: "//app:cli",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeBazel {
		t.Fatalf("expected mode bazel, got %s", resolved.Mode)
	}
	if resolved.BazelWorkspacePath != dir {
		t.Fatalf("expected bazel workspace path %s, got %s", dir, resolved.BazelWorkspacePath)
	}
	if resolved.BazelTargets != "//app:cli" {
		t.Fatalf("expected bazel targets //app:cli, got %s", resolved.BazelTargets)
	}
}

func TestResolveXcodeWorkspaceViaContents(t *testing.T) {
	dir := t.TempDir()
	workspace := filepath.Join(dir, "App.xcworkspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	project := filepath.Join(dir, "App.xcodeproj")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	contents := `<?xml version="1.0" encoding="UTF-8"?>
<Workspace>
  <FileRef location = "group:App.xcodeproj"></FileRef>
</Workspace>`
	if err := os.WriteFile(filepath.Join(workspace, "contents.xcworkspacedata"), []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write workspace contents: %v", err)
	}

	resolved, err := Resolve(Request{Path: dir, Mode: ModeAuto})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeXcode {
		t.Fatalf("expected xcode mode, got %s", resolved.Mode)
	}
	if resolved.WorkspacePath != workspace {
		t.Fatalf("expected workspace path %s, got %s", workspace, resolved.WorkspacePath)
	}
	if resolved.ProjectPath != project {
		t.Fatalf("expected project path %s, got %s", project, resolved.ProjectPath)
	}
}

func TestResolveAutoNoInputFound(t *testing.T) {
	_, err := Resolve(Request{Path: t.TempDir(), Mode: ModeAuto})
	if err == nil {
		t.Fatal("expected input not found error")
	}
	if !apperrors.IsKind(err, apperrors.KindInputNotFound) {
		t.Fatalf("expected input not found kind, got %v", err)
	}
	if !strings.Contains(err.Error(), "no supported project markers found under") {
		t.Fatalf("expected combined auto mode message, got %q", err.Error())
	}
	if strings.Contains(err.Error(), "Try --mode spm|xcode|bazel") {
		t.Fatalf("did not expect mode hint in error, got %q", err.Error())
	}
}

func TestResolveAutoWithExplicitMissingProjectReturnsXcodeError(t *testing.T) {
	_, err := Resolve(Request{
		Path:        t.TempDir(),
		Mode:        ModeAuto,
		ProjectPath: filepath.Join(t.TempDir(), "Missing.xcodeproj"),
	})
	if err == nil {
		t.Fatal("expected xcode project not found error")
	}
	if !apperrors.IsKind(err, apperrors.KindXcodeProjectNotFound) {
		t.Fatalf("expected xcode project not found kind, got %v", err)
	}
}

func TestResolveSPMIgnoresXcodeFlags(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// test"), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	resolved, err := Resolve(Request{
		Path:        dir,
		Mode:        ModeSPM,
		ProjectPath: filepath.Join(dir, "App.xcodeproj"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Mode != ModeSPM {
		t.Fatalf("expected mode spm, got %s", resolved.Mode)
	}
	if resolved.PackagePath != dir {
		t.Fatalf("unexpected package path %s", resolved.PackagePath)
	}
}

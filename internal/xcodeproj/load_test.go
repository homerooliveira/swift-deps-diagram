package xcodeproj

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	apperrors "swift-deps-diagram/internal/errors"
)

func TestLoadParsesTargetsAndProducts(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "App.xcodeproj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "project.pbxproj"), []byte("// placeholder"), 0o644); err != nil {
		t.Fatalf("failed to create pbxproj: %v", err)
	}

	oldLookPath := lookPath
	oldRunPlutil := runPlutil
	lookPath = func(string) (string, error) { return "/usr/bin/plutil", nil }
	runPlutil = func(context.Context, string) ([]byte, []byte, error) {
		json := `{
  "objects": {
    "TARGET_APP": {
      "isa": "PBXNativeTarget",
      "name": "App",
      "dependencies": ["DEP_CORE"],
      "packageProductDependencies": ["PROD_ALAMOFIRE"]
    },
    "TARGET_CORE": {
      "isa": "PBXNativeTarget",
      "name": "Core"
    },
    "DEP_CORE": {
      "isa": "PBXTargetDependency",
      "target": "TARGET_CORE"
    },
    "PKG_REMOTE": {
      "isa": "XCRemoteSwiftPackageReference",
      "identity": "alamofire"
    },
    "PROD_ALAMOFIRE": {
      "isa": "XCSwiftPackageProductDependency",
      "productName": "Alamofire",
      "package": "PKG_REMOTE"
    }
  }
}`
		return []byte(json), nil, nil
	}
	t.Cleanup(func() {
		lookPath = oldLookPath
		runPlutil = oldRunPlutil
	})

	project, err := Load(context.Background(), projectDir)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if len(project.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(project.Targets))
	}

	var appTarget Target
	for _, target := range project.Targets {
		if target.Name == "App" {
			appTarget = target
		}
	}
	if appTarget.Name != "App" {
		t.Fatal("expected App target")
	}
	if len(appTarget.TargetDependsOn) != 1 || appTarget.TargetDependsOn[0] != "TARGET_CORE" {
		t.Fatalf("unexpected target deps: %#v", appTarget.TargetDependsOn)
	}
	if len(appTarget.Products) != 1 || appTarget.Products[0].PackageIdentity != "alamofire" {
		t.Fatalf("unexpected product deps: %#v", appTarget.Products)
	}
}

func TestLoadPlutilFailure(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "App.xcodeproj")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "project.pbxproj"), []byte("// placeholder"), 0o644); err != nil {
		t.Fatalf("failed to create pbxproj: %v", err)
	}

	oldLookPath := lookPath
	oldRunPlutil := runPlutil
	lookPath = func(string) (string, error) { return "/usr/bin/plutil", nil }
	runPlutil = func(context.Context, string) ([]byte, []byte, error) {
		return nil, []byte("broken file"), errors.New("exit 1")
	}
	t.Cleanup(func() {
		lookPath = oldLookPath
		runPlutil = oldRunPlutil
	})

	_, err := Load(context.Background(), projectDir)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !apperrors.IsKind(err, apperrors.KindXcodeParse) {
		t.Fatalf("expected xcode parse kind, got %v", err)
	}
}

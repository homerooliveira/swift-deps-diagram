package manifest

import (
	"testing"

	"swift-deps-diagram/internal/testutil"
)

func TestDecodeManifestMinimal(t *testing.T) {
	data := testutil.ReadFixture(t, "simple-local.json")
	pkg, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if pkg.Name != "Sample" {
		t.Fatalf("expected package Sample, got %q", pkg.Name)
	}
	if len(pkg.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(pkg.Targets))
	}
}

func TestDecodeManifestDependenciesVariants(t *testing.T) {
	data := testutil.ReadFixture(t, "product-and-byname.json")
	pkg, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	deps := pkg.Targets[0].Dependencies
	if len(deps) < 3 {
		t.Fatalf("expected at least 3 dependencies, got %d", len(deps))
	}
	if deps[0].Kind != DependencyKindProduct {
		t.Fatalf("expected first dep to be product, got %s", deps[0].Kind)
	}
	if deps[1].Kind != DependencyKindByName {
		t.Fatalf("expected second dep to be byName, got %s", deps[1].Kind)
	}
}

func TestDecodeManifestUnknownFieldsIgnored(t *testing.T) {
	data := []byte(`{"name":"Pkg","targets":[],"unknown":{"x":1}}`)
	pkg, err := Decode(data)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if pkg.Name != "Pkg" {
		t.Fatalf("expected Pkg, got %q", pkg.Name)
	}
}

func TestDecodeManifestInvalidJSON(t *testing.T) {
	data := testutil.ReadFixture(t, "malformed.json")
	_, err := Decode(data)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

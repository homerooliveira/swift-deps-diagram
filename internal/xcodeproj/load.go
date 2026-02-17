package xcodeproj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	apperrors "swift-deps-diagram/internal/errors"
)

type Project struct {
	Targets []Target
}

type Target struct {
	ID              string
	Name            string
	ProductType     string
	TargetDependsOn []string
	Products        []PackageProduct
}

type PackageProduct struct {
	Name            string
	PackageIdentity string
}

const parseTimeout = 30 * time.Second

var lookPath = exec.LookPath

var runPlutil = func(ctx context.Context, pbxprojPath string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, "plutil", "-convert", "json", "-o", "-", pbxprojPath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

type pbxRoot struct {
	Objects map[string]map[string]interface{} `json:"objects"`
}

// Load parses an .xcodeproj and returns target + SPM product dependencies.
func Load(ctx context.Context, xcodeprojPath string) (Project, error) {
	if _, err := lookPath("plutil"); err != nil {
		return Project{}, apperrors.New(apperrors.KindXcodeParse, "plutil binary not found in PATH", err)
	}
	if !strings.HasSuffix(xcodeprojPath, ".xcodeproj") {
		return Project{}, apperrors.New(apperrors.KindXcodeProjectNotFound, "xcode project path must end with .xcodeproj", nil)
	}

	if _, err := os.Stat(xcodeprojPath); err != nil {
		return Project{}, apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("xcode project not found at %s", xcodeprojPath), err)
	}

	pbxprojPath := filepath.Join(xcodeprojPath, "project.pbxproj")
	if _, err := os.Stat(pbxprojPath); err != nil {
		return Project{}, apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("project.pbxproj not found at %s", pbxprojPath), err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, parseTimeout)
	defer cancel()

	stdout, stderr, err := runPlutil(timeoutCtx, pbxprojPath)
	if err != nil {
		detail := strings.TrimSpace(string(stderr))
		if detail == "" {
			detail = err.Error()
		}
		return Project{}, apperrors.New(apperrors.KindXcodeParse, fmt.Sprintf("failed to parse project.pbxproj: %s", detail), err)
	}

	var root pbxRoot
	if err := json.Unmarshal(stdout, &root); err != nil {
		return Project{}, apperrors.New(apperrors.KindXcodeParse, "failed to decode plutil JSON output", err)
	}

	return projectFromObjects(root.Objects), nil
}

func projectFromObjects(objects map[string]map[string]interface{}) Project {
	if objects == nil {
		return Project{}
	}

	targetDeps := make(map[string]string)
	targetProxies := make(map[string]string)
	for id, obj := range objects {
		switch asString(obj["isa"]) {
		case "PBXTargetDependency":
			targetID := asString(obj["target"])
			if targetID != "" {
				targetDeps[id] = targetID
			}
			proxyID := asString(obj["targetProxy"])
			if proxyID != "" {
				targetProxies[id] = proxyID
			}
		}
	}

	proxyRemotes := make(map[string]string)
	for id, obj := range objects {
		if asString(obj["isa"]) != "PBXContainerItemProxy" {
			continue
		}
		remote := asString(obj["remoteGlobalIDString"])
		if remote != "" {
			proxyRemotes[id] = remote
		}
	}

	packageRefs := make(map[string]string)
	for id, obj := range objects {
		switch asString(obj["isa"]) {
		case "XCRemoteSwiftPackageReference":
			identity := asString(obj["identity"])
			if identity == "" {
				identity = identityFromRepositoryURL(asString(obj["repositoryURL"]))
			}
			packageRefs[id] = identity
		case "XCLocalSwiftPackageReference":
			localPath := asString(obj["relativePath"])
			if localPath == "" {
				localPath = asString(obj["path"])
			}
			if localPath == "" {
				localPath = id
			}
			packageRefs[id] = filepath.Base(localPath)
		}
	}

	productDeps := make(map[string]PackageProduct)
	for id, obj := range objects {
		if asString(obj["isa"]) != "XCSwiftPackageProductDependency" {
			continue
		}
		name := asString(obj["productName"])
		packageRef := asString(obj["package"])
		productDeps[id] = PackageProduct{Name: name, PackageIdentity: packageRefs[packageRef]}
	}

	targets := make([]Target, 0)
	for id, obj := range objects {
		isa := asString(obj["isa"])
		if isa != "PBXNativeTarget" && isa != "PBXAggregateTarget" && isa != "PBXLegacyTarget" {
			continue
		}

		t := Target{
			ID:          id,
			Name:        asString(obj["name"]),
			ProductType: asString(obj["productType"]),
		}

		for _, depID := range asStringSlice(obj["dependencies"]) {
			if targetID, ok := targetDeps[depID]; ok && targetID != "" {
				t.TargetDependsOn = append(t.TargetDependsOn, targetID)
				continue
			}
			if proxyID, ok := targetProxies[depID]; ok {
				if remoteTargetID, ok := proxyRemotes[proxyID]; ok && remoteTargetID != "" {
					t.TargetDependsOn = append(t.TargetDependsOn, remoteTargetID)
				}
			}
		}

		for _, productDepID := range asStringSlice(obj["packageProductDependencies"]) {
			if dep, ok := productDeps[productDepID]; ok && dep.Name != "" {
				t.Products = append(t.Products, dep)
			}
		}

		targets = append(targets, t)
	}

	return Project{Targets: targets}
}

func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func asStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

func identityFromRepositoryURL(repoURL string) string {
	if repoURL == "" {
		return ""
	}
	if u, err := url.Parse(repoURL); err == nil && u.Path != "" {
		base := path.Base(u.Path)
		return strings.TrimSuffix(base, ".git")
	}
	repoURL = strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repoURL, "/")
	if len(parts) == 0 {
		return repoURL
	}
	return parts[len(parts)-1]
}

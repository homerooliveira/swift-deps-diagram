package inputresolve

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	apperrors "swift-deps-diagram/internal/errors"
)

type Mode string

const (
	ModeAuto  Mode = "auto"
	ModeSPM   Mode = "spm"
	ModeXcode Mode = "xcode"
)

type Request struct {
	Path          string
	Mode          Mode
	ProjectPath   string
	WorkspacePath string
}

type Resolved struct {
	Mode          Mode
	PackagePath   string
	ProjectPath   string
	WorkspacePath string
}

func IsValidMode(mode Mode) bool {
	switch mode {
	case ModeAuto, ModeSPM, ModeXcode:
		return true
	default:
		return false
	}
}

func Resolve(req Request) (Resolved, error) {
	if req.Path == "" {
		req.Path = "."
	}
	if req.Mode == "" {
		req.Mode = ModeAuto
	}
	if !IsValidMode(req.Mode) {
		return Resolved{}, apperrors.New(apperrors.KindInvalidArgs, "--mode must be one of: auto|spm|xcode", nil)
	}
	if req.ProjectPath != "" && req.WorkspacePath != "" {
		return Resolved{}, apperrors.New(apperrors.KindInvalidArgs, "--project and --workspace cannot be used together", nil)
	}

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return Resolved{}, apperrors.New(apperrors.KindInputNotFound, "failed to resolve input path", err)
	}

	switch req.Mode {
	case ModeSPM:
		pkgPath, err := resolvePackagePath(absPath)
		if err != nil {
			return Resolved{}, err
		}
		return Resolved{Mode: ModeSPM, PackagePath: pkgPath}, nil
	case ModeXcode:
		projectPath, workspacePath, err := resolveXcodePath(absPath, req.ProjectPath, req.WorkspacePath)
		if err != nil {
			return Resolved{}, err
		}
		return Resolved{Mode: ModeXcode, ProjectPath: projectPath, WorkspacePath: workspacePath}, nil
	case ModeAuto:
		if req.ProjectPath != "" || req.WorkspacePath != "" {
			projectPath, workspacePath, err := resolveXcodePath(absPath, req.ProjectPath, req.WorkspacePath)
			if err != nil {
				return Resolved{}, err
			}
			return Resolved{Mode: ModeXcode, ProjectPath: projectPath, WorkspacePath: workspacePath}, nil
		}
		projectPath, workspacePath, err := resolveXcodePath(absPath, req.ProjectPath, req.WorkspacePath)
		if err == nil {
			return Resolved{Mode: ModeXcode, ProjectPath: projectPath, WorkspacePath: workspacePath}, nil
		}

		pkgPath, pkgErr := resolvePackagePath(absPath)
		if pkgErr == nil {
			return Resolved{Mode: ModeSPM, PackagePath: pkgPath}, nil
		}
		var appErr *apperrors.Error
		if errors.As(pkgErr, &appErr) && appErr.Kind != apperrors.KindManifestNotFound {
			return Resolved{}, pkgErr
		}
		return Resolved{}, apperrors.New(
			apperrors.KindInputNotFound,
			fmt.Sprintf("no .xcworkspace/.xcodeproj or Package.swift found under %s", absPath),
			pkgErr,
		)
	default:
		return Resolved{}, apperrors.New(apperrors.KindInvalidArgs, "unsupported mode", nil)
	}
}

func resolvePackagePath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", apperrors.New(apperrors.KindInputNotFound, "input path not found", err)
	}

	if !info.IsDir() {
		if filepath.Base(path) == "Package.swift" {
			return filepath.Dir(path), nil
		}
		return "", apperrors.New(apperrors.KindManifestNotFound, "Package.swift not found", nil)
	}

	manifestPath := filepath.Join(path, "Package.swift")
	if _, err := os.Stat(manifestPath); err != nil {
		return "", apperrors.New(apperrors.KindManifestNotFound, fmt.Sprintf("Package.swift not found at %s", manifestPath), err)
	}
	return path, nil
}

func resolveXcodePath(basePath, projectFlagPath, workspaceFlagPath string) (projectPath, workspacePath string, err error) {
	if projectFlagPath != "" {
		projectPath, err := filepath.Abs(projectFlagPath)
		if err != nil {
			return "", "", apperrors.New(apperrors.KindXcodeProjectNotFound, "failed to resolve --project path", err)
		}
		if _, statErr := os.Stat(projectPath); statErr != nil {
			return "", "", apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("xcode project not found at %s", projectPath), statErr)
		}
		return projectPath, "", nil
	}

	if workspaceFlagPath != "" {
		workspacePath, err := filepath.Abs(workspaceFlagPath)
		if err != nil {
			return "", "", apperrors.New(apperrors.KindXcodeProjectNotFound, "failed to resolve --workspace path", err)
		}
		projectPath, err := findProjectForWorkspace(workspacePath)
		if err != nil {
			return "", "", err
		}
		return projectPath, workspacePath, nil
	}

	info, statErr := os.Stat(basePath)
	if statErr != nil {
		return "", "", apperrors.New(apperrors.KindInputNotFound, "input path not found", statErr)
	}

	if !info.IsDir() {
		switch filepath.Ext(basePath) {
		case ".xcodeproj":
			return basePath, "", nil
		case ".xcworkspace":
			projectPath, err := findProjectForWorkspace(basePath)
			if err != nil {
				return "", "", err
			}
			return projectPath, basePath, nil
		default:
			return "", "", apperrors.New(apperrors.KindXcodeProjectNotFound, "xcode project or workspace not found", nil)
		}
	}

	workspaceMatches, err := filepath.Glob(filepath.Join(basePath, "*.xcworkspace"))
	if err == nil && len(workspaceMatches) > 0 {
		sort.Strings(workspaceMatches)
		workspacePath = workspaceMatches[0]
		projectPath, err = findProjectForWorkspace(workspacePath)
		if err != nil {
			return "", "", err
		}
		return projectPath, workspacePath, nil
	}

	projectMatches, err := filepath.Glob(filepath.Join(basePath, "*.xcodeproj"))
	if err == nil && len(projectMatches) > 0 {
		sort.Strings(projectMatches)
		return projectMatches[0], "", nil
	}

	return "", "", apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("no .xcworkspace/.xcodeproj found in %s", basePath), nil)
}

var workspaceRefRe = regexp.MustCompile(`location\s*=\s*"([^"]+\.xcodeproj)"`)

func findProjectForWorkspace(workspacePath string) (string, error) {
	if _, err := os.Stat(workspacePath); err != nil {
		return "", apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("workspace not found at %s", workspacePath), err)
	}

	contentsPath := filepath.Join(workspacePath, "contents.xcworkspacedata")
	data, err := os.ReadFile(contentsPath)
	if err == nil {
		matches := workspaceRefRe.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			loc := strings.TrimSpace(m[1])
			loc = strings.TrimPrefix(loc, "group:")
			loc = strings.TrimPrefix(loc, "container:")
			loc = strings.TrimPrefix(loc, "self:")
			if strings.HasPrefix(loc, "absolute:") {
				candidate := strings.TrimPrefix(loc, "absolute:")
				if _, statErr := os.Stat(candidate); statErr == nil {
					return candidate, nil
				}
				continue
			}
			candidate := filepath.Clean(filepath.Join(filepath.Dir(workspacePath), loc))
			if _, statErr := os.Stat(candidate); statErr == nil {
				return candidate, nil
			}
		}
	}

	parent := filepath.Dir(workspacePath)
	projects, globErr := filepath.Glob(filepath.Join(parent, "*.xcodeproj"))
	if globErr == nil && len(projects) > 0 {
		sort.Strings(projects)
		return projects[0], nil
	}

	return "", apperrors.New(apperrors.KindXcodeProjectNotFound, fmt.Sprintf("no .xcodeproj found for workspace %s", workspacePath), nil)
}

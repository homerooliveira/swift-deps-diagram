package apperrors

import (
	"errors"
	"fmt"
)

// Kind classifies user-facing and runtime failures.
type Kind string

const (
	KindInvalidArgs               Kind = "invalid_args"
	KindManifestNotFound          Kind = "manifest_not_found"
	KindInputNotFound             Kind = "input_not_found"
	KindAmbiguousInput            Kind = "ambiguous_input"
	KindSwiftNotFound             Kind = "swift_not_found"
	KindDumpPackage               Kind = "dump_package_failed"
	KindManifestDecode            Kind = "manifest_decode_failed"
	KindXcodeProjectNotFound      Kind = "xcode_project_not_found"
	KindXcodeParse                Kind = "xcode_parse_failed"
	KindXcodeUnsupportedStructure Kind = "xcode_unsupported_structure"
	KindBazelWorkspaceNotFound    Kind = "bazel_workspace_not_found"
	KindBazelBinaryNotFound       Kind = "bazel_binary_not_found"
	KindBazelQueryFailed          Kind = "bazel_query_failed"
	KindBazelParseFailed          Kind = "bazel_parse_failed"
	KindGraphvizNotFound          Kind = "graphviz_not_found"
	KindGraphvizRender            Kind = "graphviz_render_failed"
	KindOutputWrite               Kind = "output_write_failed"
	KindRuntime                   Kind = "runtime_failed"
)

// Error wraps typed failures so callers can map to exit codes.
type Error struct {
	Kind Kind
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg == "" && e.Err != nil {
		return e.Err.Error()
	}
	if e.Err == nil {
		return e.Msg
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(kind Kind, msg string, err error) *Error {
	return &Error{Kind: kind, Msg: msg, Err: err}
}

func IsKind(err error, kind Kind) bool {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Kind == kind
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var appErr *Error
	if !errors.As(err, &appErr) {
		return 2
	}
	switch appErr.Kind {
	case KindInvalidArgs, KindManifestNotFound, KindInputNotFound, KindAmbiguousInput, KindXcodeProjectNotFound, KindBazelWorkspaceNotFound:
		return 1
	default:
		return 2
	}
}

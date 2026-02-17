package bazel

type Target struct {
	Label string
	Kind  string
	Deps  []string
}

type Workspace struct {
	Path    string
	Scope   string
	Targets []Target
}

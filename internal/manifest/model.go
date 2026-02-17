package manifest

import "encoding/json"

// Package captures the subset of swift package dump-package needed by this CLI.
type Package struct {
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

type Target struct {
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Dependencies []TargetDependency `json:"dependencies"`
}

type DependencyKind string

const (
	DependencyKindUnknown DependencyKind = "unknown"
	DependencyKindTarget  DependencyKind = "target"
	DependencyKindProduct DependencyKind = "product"
	DependencyKindByName  DependencyKind = "by_name"
)

type TargetDependency struct {
	Kind    DependencyKind
	Name    string
	Package string
}

func (d *TargetDependency) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["target"]; ok {
		d.Kind = DependencyKindTarget
		d.Name = parseSingleName(v)
		return nil
	}
	if v, ok := raw["product"]; ok {
		d.Kind = DependencyKindProduct
		d.Name, d.Package = parseProduct(v)
		return nil
	}
	if v, ok := raw["byName"]; ok {
		d.Kind = DependencyKindByName
		d.Name = parseSingleName(v)
		return nil
	}

	d.Kind = DependencyKindUnknown
	return nil
}

func parseSingleName(raw json.RawMessage) string {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}

	var asArray []interface{}
	if err := json.Unmarshal(raw, &asArray); err == nil {
		for _, item := range asArray {
			if s, ok := item.(string); ok && s != "" {
				return s
			}
		}
	}

	var asObject map[string]interface{}
	if err := json.Unmarshal(raw, &asObject); err == nil {
		if name, ok := asObject["name"].(string); ok {
			return name
		}
	}

	return ""
}

func parseProduct(raw json.RawMessage) (name, pkg string) {
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString, ""
	}

	var asArray []interface{}
	if err := json.Unmarshal(raw, &asArray); err == nil {
		if len(asArray) > 0 {
			if s, ok := asArray[0].(string); ok {
				name = s
			}
		}
		if len(asArray) > 1 {
			if s, ok := asArray[1].(string); ok {
				pkg = s
			}
		}
		return name, pkg
	}

	var asObject map[string]interface{}
	if err := json.Unmarshal(raw, &asObject); err == nil {
		if s, ok := asObject["name"].(string); ok {
			name = s
		}
		if s, ok := asObject["package"].(string); ok {
			pkg = s
		}
	}

	return name, pkg
}

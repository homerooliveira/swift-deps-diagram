package manifest

import (
	"encoding/json"

	apperrors "swift-deps-diagram/internal/errors"
)

// Decode transforms dump-package JSON into typed manifest data.
func Decode(data []byte) (Package, error) {
	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Package{}, apperrors.New(apperrors.KindManifestDecode, "failed to decode dump-package output", err)
	}
	return pkg, nil
}

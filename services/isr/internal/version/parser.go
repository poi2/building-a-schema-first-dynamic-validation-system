package version

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSemVer parses a semantic version string (e.g., "1.2.3") into major, minor, patch components.
// Returns an error if the version format is invalid or contains non-numeric components.
func ParseSemVer(version string) (major, minor, patch int32, err error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: expected 3 parts, got %d", len(parts))
	}

	majorInt, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minorInt, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	patchInt, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
	}

	if majorInt < 0 || minorInt < 0 || patchInt < 0 {
		return 0, 0, 0, fmt.Errorf("version components must be non-negative")
	}

	return int32(majorInt), int32(minorInt), int32(patchInt), nil
}

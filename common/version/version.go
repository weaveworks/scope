package version

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	versionRegexp = "^v?(\\d+)\\.(\\d+)\\.(\\d+)$"
)

// Version represents a release version decomposed in (major,minor,patch)
type Version struct {
	major int
	minor int
	patch int
}

// New returns a new Version
func New(major, minor, patch int) Version {
	return Version{major, minor, patch}
}

// ParseVersion parses a version string and, if successful, returns a Version
func ParseVersion(version string) (*Version, error) {
	// the regexp catches this, short circuiting does not hurt though
	if version == "" {
		return nil, fmt.Errorf("Invalid version: %v", version)
	}

	re := regexp.MustCompile(versionRegexp)
	match := re.FindStringSubmatch(version)
	if len(match) != 4 {
		return nil, fmt.Errorf("Invalid version: %v", version)
	}

	major, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid version: %v", version)
	}
	minor, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, fmt.Errorf("Invalid version: %v", version)
	}
	patch, err := strconv.Atoi(match[3])
	if err != nil {
		return nil, fmt.Errorf("Invalid version: %v", version)
	}

	return &Version{major, minor, patch}, nil
}

// Lower compares v against version and returns true if v is lower than version
func (v Version) Lower(version Version) bool {
	if v.major < version.major {
		return true
	}
	if v.major > version.major {
		return false
	}
	// major versions are equal
	if v.minor < version.minor {
		return true
	}
	if v.minor > version.minor {
		return false
	}
	// minor versions are equal
	return v.patch < version.patch
}

func (v Version) String() string {
	return fmt.Sprintf("%v.%v.%v", v.major, v.minor, v.patch)
}

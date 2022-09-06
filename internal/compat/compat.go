package compat

import (
	"golang.org/x/mod/semver"
)

// this is the minimum supported version of Lagoon that the build tool can be used with
// if there is a change in lagoon-core that the build tool has to support, then this value
// should be adjusted to the version of lagoon-core to support
// this could also be modified with a compiler flag for developing
var supportedMinVersion = "v2.9.0"

func checkVersion(version, minVersion string) bool {
	comp := semver.Compare(version, minVersion)
	if comp == -1 {
		return false
	}
	return true
}

func CheckVersion(version string) bool {
	return checkVersion(version, supportedMinVersion)
}

func SupportedMinVersion() string {
	return supportedMinVersion
}

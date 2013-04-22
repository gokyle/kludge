package common

import "fmt"

var version struct {
	Major byte
	Minor byte
	Patch uint16
}

func init() {
	version.Major = 0
	version.Minor = 1
	version.Patch = 3
}

func Version() string {
	return fmt.Sprintf("kludge-%d.%d.%d", version.Major,
		version.Minor, version.Patch)
}

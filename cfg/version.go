package cfg

import (
	"regexp"
	"runtime/debug"
)

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build info")
	}

	if info.Main.Version != "" {
		return versionRegex.ReplaceAllString(info.Main.Version, "")
	}

	return "(dev build)"
}

// Ignore timestamp and git hash
var versionRegex = regexp.MustCompile(`-0\.\d{14}-[0-9a-f]+(\+dirty)?`)

package cfg

import (
	"runtime/debug"
)

func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build info")
	}
	if info.Main.Version != "" {
		return info.Main.Version
	}

	return "(devel)"
}

package module

import (
	"runtime/debug"
	"sync"
)

// Path returns the path of the main module.
func Path() string {
	build := buildInfo()
	if build == nil {
		return "(none)"
	}
	return build.Main.Path
}

// Version returns the version string of the main module.
func Version() string {
	return versionString()
}

var buildInfo = sync.OnceValue(func() *debug.BuildInfo {
	build, _ := debug.ReadBuildInfo()
	return build
})

var versionString = sync.OnceValue(func() string {
	build := buildInfo()
	if build == nil {
		return "(none)"
	}
	version := build.Main.Version
	var revision string
	for _, s := range build.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		}
	}
	if version == "" {
		version = "(none)"
	}
	versionString := version
	if revision != "" {
		versionString += " (" + revision + ")"
	}
	return versionString
})

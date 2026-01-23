// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application.
	// This will be set at build time using ldflags.
	Version = "dev"

	// BuildTime is the time when the binary was built.
	// This will be set at build time using ldflags.
	BuildTime = "unknown"

	// GitCommit is the git commit hash when the binary was built.
	// This will be set at build time using ldflags.
	GitCommit = "unknown"
)

// Info holds version and build information.
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns version information.
func Get() Info {
	return Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string.
func (i Info) String() string {
	return fmt.Sprintf("Version: %s, BuildTime: %s, GitCommit: %s, GoVersion: %s, Platform: %s",
		i.Version, i.BuildTime, i.GitCommit, i.GoVersion, i.Platform)
}

// Short returns a short version string.
func (i Info) Short() string {
	if i.Version == "dev" {
		return fmt.Sprintf("%s (dev build)", i.Version)
	}
	return i.Version
}

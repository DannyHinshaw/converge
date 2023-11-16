package version

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

const (
	gitRevShortLen  = 12
	fallbackVersion = "(devel)"
)

// GetVersion retrieves the version of the module dynamically at runtime.
func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return getGitVersion()
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs" && setting.Value == "git" {
			return getGitVersion()
		}
	}

	return constructPseudoVersion(info.Settings)
}

func getGitVersion() string {
	// Try to get the latest tag
	tag, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err == nil && len(tag) > 0 {
		return strings.TrimSpace(string(tag))
	}

	// Fallback to the latest commit hash
	commit, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err == nil && len(commit) > 0 {
		commitHash := strings.TrimSpace(string(commit))
		if len(commitHash) > gitRevShortLen {
			commitHash = commitHash[:gitRevShortLen]
		}
		return fmt.Sprintf("commit-%s", commitHash)
	}

	return fallbackVersion
}

func constructPseudoVersion(bs []debug.BuildSetting) string {
	var vcsTime time.Time
	var vcsRev string
	for _, s := range bs {
		switch s.Key {
		case "vcs.time":
			vcsTime, _ = time.Parse(time.RFC3339Nano, s.Value)
		case "vcs.revision":
			vcsRev = s.Value
		}
	}

	// Format the timestamp in the specific format used by Go pseudo-versions
	timestamp := vcsTime.UTC().Format("20060102150405")

	// Truncate the commit hash to the first 12 characters
	if len(vcsRev) > gitRevShortLen {
		vcsRev = vcsRev[:gitRevShortLen]
	}

	if vcsRev != "" {
		return fmt.Sprintf("v0.0.0-%s-%s", timestamp, vcsRev)
	}

	return fallbackVersion
}

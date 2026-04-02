package codexmgr

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const MinVersion = "0.117.0"

type semver struct {
	Major int
	Minor int
	Patch int
}

var semverRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func parseSemver(s string) (*semver, error) {
	m := semverRe.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("cannot parse version from %q", s)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return &semver{major, minor, patch}, nil
}

func (v *semver) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *semver) lessThan(other *semver) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}

func CheckVersion() error {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return fmt.Errorf("codex binary not found in PATH")
	}

	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get codex version: %w", err)
	}

	current, err := parseSemver(strings.TrimSpace(string(out)))
	if err != nil {
		return fmt.Errorf("cannot parse codex version output: %s", strings.TrimSpace(string(out)))
	}

	min, _ := parseSemver(MinVersion)

	if current.lessThan(min) {
		return fmt.Errorf("codex-cli %s is below minimum required %s", current, min)
	}
	return nil
}

func DetectedVersion() (string, error) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return "", fmt.Errorf("codex not found in PATH")
	}
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		return "", err
	}
	v, err := parseSemver(strings.TrimSpace(string(out)))
	if err != nil {
		return strings.TrimSpace(string(out)), nil
	}
	return v.String(), nil
}

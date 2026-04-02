package gitutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\n%s", err, hintFromOutput(string(out)))
	}
	return nil
}

func Pull(repoDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "pull")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s\n%s", err, hintFromOutput(string(out)))
	}
	return nil
}

func Checkout(repoDir, ref string) error {
	cmd := exec.Command("git", "-C", repoDir, "checkout", ref)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout %s failed: %s\n%s", ref, err, hintFromOutput(string(out)))
	}
	return nil
}

func CurrentCommitSHA(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func CurrentBranch(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func RemoteURL(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func IsGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func ListTags(repoDir string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoDir, "tag", "--sort=-v:refname")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func hintFromOutput(output string) string {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "could not resolve host"):
		return "Hint: check network / VPN / DNS access to the Git host."
	case strings.Contains(lower, "authentication failed"),
		strings.Contains(lower, "could not read username"),
		strings.Contains(lower, "access denied"):
		return "Hint: this Git URL likely needs credentials or a configured git credential helper."
	default:
		return strings.TrimSpace(output)
	}
}

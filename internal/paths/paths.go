package paths

import (
	"os"
	"path/filepath"
)

const (
	ConfigDirName = ".osk"
	BinaryName    = "openskills"
)

func Home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return h
}

func ConfigDir() string {
	if v := os.Getenv("OSK_HOME"); v != "" {
		return v
	}
	return filepath.Join(Home(), ConfigDirName)
}

func ConfigFile() string { return filepath.Join(ConfigDir(), "config.toml") }
func StateFile() string  { return filepath.Join(ConfigDir(), "state.json") }
func LockFile() string   { return filepath.Join(ConfigDir(), "openskills.lock") }
func ReposDir() string   { return filepath.Join(ConfigDir(), "repos") }
func SkillsDir() string  { return filepath.Join(ConfigDir(), "skills") }
func BackupsDir() string { return filepath.Join(ConfigDir(), "backups") }

func CodexHome() string {
	if v := os.Getenv("CODEX_HOME"); v != "" {
		return v
	}
	return filepath.Join(Home(), ".codex")
}

func ClaudeHome() string {
	return filepath.Join(Home(), ".claude")
}

func CursorHome() string {
	return filepath.Join(Home(), ".cursor")
}

func AgentsDir() string {
	return filepath.Join(Home(), ".agents")
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

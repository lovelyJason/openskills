package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lovelyJason/openskills/internal/paths"
)

type Entry struct {
	OriginalPath  string `json:"originalPath"`
	BackupPath    string `json:"backupPath"`
	WasSymlink    bool   `json:"wasSymlink,omitempty"`
	SymlinkTarget string `json:"symlinkTarget,omitempty"`
	WasAbsent     bool   `json:"wasAbsent,omitempty"`
}

type Manifest struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Operation string    `json:"operation"`
	Entries   []Entry   `json:"entries"`
}

type Transaction struct {
	manifest Manifest
	dir      string
	filesDir string
}

func Begin(operation string) (*Transaction, error) {
	id := time.Now().Format("20060102T150405")
	dir := filepath.Join(paths.BackupsDir(), id)
	filesDir := filepath.Join(dir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return nil, err
	}
	return &Transaction{
		manifest: Manifest{
			ID:        id,
			CreatedAt: time.Now(),
			Operation: operation,
		},
		dir:      dir,
		filesDir: filesDir,
	}, nil
}

func (tx *Transaction) BackupFile(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		tx.manifest.Entries = append(tx.manifest.Entries, Entry{
			OriginalPath: path,
			WasAbsent:    true,
		})
		return nil
	}
	if err != nil {
		return err
	}

	safeName := strings.ReplaceAll(path, "/", "__")
	backupPath := filepath.Join(tx.filesDir, safeName)

	entry := Entry{
		OriginalPath: path,
		BackupPath:   backupPath,
	}

	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		entry.WasSymlink = true
		entry.SymlinkTarget = target
	} else if info.IsDir() {
		if err := copyDir(path, backupPath); err != nil {
			return err
		}
	} else {
		if err := copyFile(path, backupPath); err != nil {
			return err
		}
	}

	tx.manifest.Entries = append(tx.manifest.Entries, entry)
	return nil
}

func (tx *Transaction) Commit() error {
	data, err := json.MarshalIndent(tx.manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(tx.dir, "manifest.json"), data, 0644)
}

func (tx *Transaction) Rollback() error {
	for i := len(tx.manifest.Entries) - 1; i >= 0; i-- {
		entry := tx.manifest.Entries[i]
		if entry.WasAbsent {
			os.RemoveAll(entry.OriginalPath)
			continue
		}
		os.RemoveAll(entry.OriginalPath)
		if entry.WasSymlink {
			os.Symlink(entry.SymlinkTarget, entry.OriginalPath)
		} else if entry.BackupPath != "" {
			info, err := os.Stat(entry.BackupPath)
			if err != nil {
				continue
			}
			if info.IsDir() {
				copyDir(entry.BackupPath, entry.OriginalPath)
			} else {
				copyFile(entry.BackupPath, entry.OriginalPath)
			}
		}
	}
	return nil
}

func (tx *Transaction) Cleanup() {
	os.RemoveAll(tx.dir)
}

func PruneBackups(maxKeep int) error {
	dir := paths.BackupsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)

	if len(dirs) <= maxKeep {
		return nil
	}
	for _, d := range dirs[:len(dirs)-maxKeep] {
		os.RemoveAll(filepath.Join(dir, d))
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		return copyFile(path, target)
	})
}

func RollbackLatest() error {
	dir := paths.BackupsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("no backups found: %w", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no backups found")
	}
	sort.Strings(dirs)
	latest := dirs[len(dirs)-1]

	data, err := os.ReadFile(filepath.Join(dir, latest, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	tx := &Transaction{
		manifest: manifest,
		dir:      filepath.Join(dir, latest),
		filesDir: filepath.Join(dir, latest, "files"),
	}
	return tx.Rollback()
}

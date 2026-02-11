package store

import (
	"os"
	"path/filepath"
	"runtime"
)

func EnsureSymlink(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	_ = os.Remove(dst)

	if err := os.Symlink(src, dst); err == nil {
		return nil
	}

	// Windows fallback when symlink permissions are restricted.
	if runtime.GOOS == "windows" {
		return copyDir(src, dst)
	}
	return nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		raw, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

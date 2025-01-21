package jerminal

import (
	"io"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)

	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func copyDir(src, dst string) error {

	_, err := os.Stat(src)

	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func isDirEmpty(dirPath string) (bool, error) {

	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Read the directory contents
	entries, err := dir.Readdir(1) // Read up to 1 entry
	if err != nil {
		return false, err
	}

	// If no entries are found, the directory is empty
	return len(entries) == 0, nil
}

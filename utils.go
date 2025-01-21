package jerminal

import (
	"io"
	"os"
)

func copyDir(src, dst string) error {

	_, err := os.Stat(src)
    if err != nil {
        return err
    }

    dir := os.DirFS(src)
    err = os.CopyFS(dst, dir)

	return err
}

func isDirEmpty(dirPath string) (bool, error) {

	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	_, err = dir.Readdir(1)

    if err == io.EOF {
        return true, nil
    }

    return false, err
}

package utils

import (
	"encoding/json"
	"io"
	"os"
)

func CopyDir(src, dst string) error {

	_, err := os.Stat(src)
    if err != nil {
        return err
    }

    dir := os.DirFS(src)
    err = os.CopyFS(dst, dir)

	return err
}

func IsDirEmpty(dirPath string) (bool, error) {

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

func MustMarshall(val any) []byte {
    bytes, err := json.Marshal(val)
    if err != nil {
        panic(err)
    }
    return bytes
}

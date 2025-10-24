package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// File permission constants
const (
	DirPerm  = 0o700 // rwx------
	FilePerm = 0o600 // rw-------
)

// WriteToFile writes data to a file, creating directories as needed.
// It creates parent directories with DirPerm and the file with FilePerm.
func WriteToFile(source []byte, file string) error {
	err := os.MkdirAll(filepath.Dir(file), DirPerm)
	if err != nil {
		return fmt.Errorf("create output directory:%w", err)
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FilePerm)
	if err != nil {
		return fmt.Errorf("open file:%w", err)
	}

	_, err = f.Write(source)

	closeErr := f.Close()
	if closeErr != nil {
		return fmt.Errorf("closing file:%w", closeErr)
	}

	if err != nil {
		return fmt.Errorf("writing file:%w", err)
	}

	return nil
}

// FileExists checks if a file exists and is a regular file (not a directory).
func FileExists(filename string) bool {
	fileInfo, err := os.Stat(filename)
	return err == nil && fileInfo.Mode().IsRegular()
}

// ChangeDirectory modifies a file path by changing its directory prefix.
// If the filename starts with swapDir, it replaces it with dir.
// Otherwise, it prepends dir to the path.
func ChangeDirectory(dir, swapDir, filename string) string {
	path := strings.Split(filename, string(os.PathSeparator))
	if len(path) == 0 {
		return filename
	}

	if path[0] == swapDir {
		path[0] = dir
	} else {
		path = append([]string{dir}, path...)
	}

	return filepath.Join(path...)
}

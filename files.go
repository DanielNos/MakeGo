package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func makeDirs(paths []string, perm os.FileMode) error {
	for _, path := range paths {
		err := os.MkdirAll(path, perm)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeLine(file *os.File, line string) {
	file.WriteString(line + "\n")
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func copyFile(from, to string) error {
	cmd := exec.Command("cp", from, to)
	_, err := cmd.CombinedOutput()
	return err
}

func getExtension(path string) string {
	splitPath := strings.Split(path, ".")
	if len(splitPath) == 1 {
		return ""
	}
	return splitPath[len(splitPath)-1]
}

func addXPerm(file string) error {
	cmd := exec.Command("chmod", "+x", file)
	_, err := cmd.CombinedOutput()
	return err
}

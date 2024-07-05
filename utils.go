package main

import (
	"fmt"
	"io"
	"os"
)

func copyFile(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file for writing
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents from source file to destination file
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	// Flush the writes to the stable storage
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

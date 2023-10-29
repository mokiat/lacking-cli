package distribution

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateBlankDistFolder(projectDirLoc, platform string) (string, error) {
	const distFolderName = "dist"

	distFolderLoc := filepath.Join(projectDirLoc, distFolderName, platform)
	if err := os.RemoveAll(distFolderLoc); err != nil {
		return "", fmt.Errorf("error removing existing folder: %w", err)
	}

	if err := os.MkdirAll(distFolderLoc, 0o775); err != nil {
		return "", fmt.Errorf("error creating new folder: %w", err)
	}

	return distFolderLoc, nil
}

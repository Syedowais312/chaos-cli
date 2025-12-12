package utils

import (
	"os"
	"path/filepath"
)

func ResolveOutputPath(filename string) (string, error) {
	//get user's current drectory

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	outputDir := filepath.Join(cwd, "chaos-cli-test")

	//checking if dir exists or not
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(outputDir, filename), nil
}

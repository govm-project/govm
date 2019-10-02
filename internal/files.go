package internal

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CheckFilePath verifies and fixes a file path
func CheckFilePath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		path = strings.Replace(path, "~", GetUserHomePath(), 1)
	}

	imgArg, err := os.Stat(path)
	if err != nil {
		return path, fmt.Errorf("file %v does not exist", path)
	}

	if pathDir := filepath.Dir(path); pathDir == "." {
		pathDir, err = os.Getwd()
		if err != nil {
			return path, fmt.Errorf("file's absolute path cannot be constructed for  %v", path)
		}
		path = fmt.Sprintf("%v/%v", pathDir, path)
	}

	mode := imgArg.Mode()
	if !mode.IsRegular() {
		return path, fmt.Errorf("%v is not a regular file", path)
	}

	return filepath.Abs(path)
}

// GetUserHomePath gets the users Home Path
func GetUserHomePath() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Warn("Unable to determine $HOME")
		log.Error("Please specify -workdir and -pubkey")
	}
	return currentUser.HomeDir
}

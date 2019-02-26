package homedir

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// HomeDir returns the home directory for the current user
func HomeDir() string {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("HOME"); home != "" {
			if _, err := os.Stat(home); err == nil {
				return home
			}
		}
		if home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH"); home != "" {
			if _, err := os.Stat(home); err == nil {
				return home
			}
		}
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			if _, err := os.Stat(userProfile); err == nil {
				return userProfile
			}
		}
	}
	return os.Getenv("HOME")
}

// ExpandPath returns the given path with environment variables expanded
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(HomeDir(), path[1:])
	}
	return os.ExpandEnv(path)
}

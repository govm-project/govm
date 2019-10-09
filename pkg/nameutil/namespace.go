package nameutil

import (
	"fmt"
	"os/user"
)

// DefaultNamespace returns the default namespace for a user, which generally is
// the user's username.
func DefaultNamespace() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get current user: %v", err)
	}

	return u.Username, nil
}

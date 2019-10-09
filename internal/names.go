package internal

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	log "github.com/sirupsen/logrus"
)

const containerPrefix = "govm"

// GenerateContainerName generates a container name based on the namespace
// and virtual machine name.
func GenerateContainerName(namespace, vmname string) string {
	return fmt.Sprintf("%s.%s.%s", containerPrefix, namespace, vmname)
}

// ParseContainerName takes a container name and returns the VM namespace and
// name, or an error if the container name is incorrectly formatted.
func ParseContainerName(contname string) (ns, vmname string, err error) {
	parts := strings.Split(contname, ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("%q has an invalid container name", contname)
	}

	if parts[0] != containerPrefix {
		return parts[1], parts[2], fmt.Errorf(
			"%q has an invalid prefix (should be %q)", contname, containerPrefix)
	}

	return parts[1], parts[2], nil
}

// DefaultNamespace returns the default namespace for a user, which generally is
// the user's username.
func DefaultNamespace() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get current user: %v", err)
	}

	return u.Username, nil
}

// RandomName returns a randomly generated name. This just wraps docker's
// namesgenerator.GetRandomName() to avoid extra imports.
func RandomName() string {
	return namesgenerator.GetRandomName(0)
}

// GetDefaultWorkDir returns the default working directory
func GetDefaultWorkDir() string {
	homeDir := GetUserHomePath()
	workDir := fmt.Sprintf("%v/vms", homeDir)
	_, err := os.Stat(workDir)

	if err != nil {
		log.WithField("workdir", workDir).Warn(
			"Work Directory does not exist")

		log.WithField("workdir", workDir+"/data").Info(
			"Creating workdir")

		err = os.MkdirAll(workDir+"/data", 0755) // nolint: gas
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Creating %s", workDir+"/images")
		err = os.Mkdir(workDir+"/images", 0755) // nolint: gas

		if err != nil {
			log.Fatal(err)
		}
	}

	return workDir
}

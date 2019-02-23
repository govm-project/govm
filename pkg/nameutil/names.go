package nameutil

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
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

// RandomName returns a randomly generated name. This just wraps docker's
// namesgenerator.GetRandomName() to avoid extra imports.
func RandomName() string {
	return namesgenerator.GetRandomName(0)
}

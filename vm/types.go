package vm

import (
	"github.com/govm-project/govm/engines/docker"
)

//ComposeTemplate defines a VMs orchestration template
type ComposeTemplate struct {
	VMs      []VM             `yaml:"vms"`
	Networks []docker.Network `yaml:"networks"`
}

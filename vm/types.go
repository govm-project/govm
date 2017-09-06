package vm

import (
	"github.intel.com/clrgdc/govm/docker"
)

//ComposeTemplate defines a VMs orchestration template
type ComposeTemplate struct {
	VMs      []VM             `yaml:"vms"`
	Networks []docker.Network `yaml:"networks"`
}

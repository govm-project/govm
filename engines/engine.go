package main

import (
	"github.com/govm-project/govm/vm"
)

// Engine stands as an abstraction for VMs management engines
type Engine interface {
	CreateVM(spec vm.VM) error
}

package main

import (
	"github.com/govm-project/govm/vm"
)

// VMEngine stands as an abstraction for VMs management engines
type VMEngine interface {
	Create(spec vm.Instance) (string, error)
	Start(namespace, id string) error
	Delete(namespace, id string) error
	List(namespace string, all bool) ([]vm.Instance, error)
}

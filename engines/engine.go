package engines

import (
	"github.com/govm-project/govm/pkg/termutil"
	"github.com/govm-project/govm/vm"
)

// VMEngine stands as an abstraction for VMs management engines
type VMEngine interface {
	CreateVM(spec vm.Instance) (string, error)
	StartVM(namespace, id string) error
	StopVM(namespace, id string) error
	DeleteVM(namespace, id string) error
	SSHVM(namespace, id, user, key string, term *termutil.Terminal) error
	ListVM(namespace string, all bool) ([]vm.Instance, error)
	SaveVM(namespace, id, outputFile string, stopVM bool) error
}

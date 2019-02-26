package types

import (
	"net"
)

// VM represents a Virtual Machine guest
type VM struct {
	Name        string    `yaml:"name" json:"name"`
	Namespace   string    `yaml:"namespace" json:"namespace"`
	ParentImage string    `yaml:"image" json:"image"`
	Size        Size      `yaml:"size" json:"size"`
	SSHKey      string    `yaml:"sshkey" json:"sshkey"`
	UserData    string    `yaml:"user-data" json:"user_data"`
	Cloud       bool      `yaml:"cloud" json:"cloud"`
	Efi         bool      `yaml:"efi" json:"efi"`
	AutoRemove  bool      `yaml:"auto-remove" json:"auto_remove"`
	Network     VMNetOpts `yaml:"network" json:"network"`
	Shares      []string  `yaml:"shares" json:"shares"`
	EmulatorEnv []string  `yaml:"emulator-env" json:"emulator_env"`
}

// VMNetOpts represents a VM's options for connecting to a network
type VMNetOpts struct {
	NetID string
	IP    net.IP
	MAC   net.HardwareAddr
}

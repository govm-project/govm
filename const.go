package main

// Container Images
const (
	VMLauncherContainerImage = "vmgo/vmgo"
	VNCContainerImage        = "vmgo/novnc-server"
)

// Default VM Launcher containers' names
const (
	VNCServerContainerName = "vm-launcher-novnc-server"
)

// VM launcher environment variables
const (
	VMLauncherWorkdir = "$HOME/vms"
	SSHPublicKeyFile  = "$HOME/.ssh/id_rsa.pub"
)

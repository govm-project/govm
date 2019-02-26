package types

// Size represents all hardware spec definitions for the VM
type Size struct {
	Name     string `yaml:"name" json:"name"`
	CPUModel string `yaml:"cpu-model" json:"cpu_model"`
	CPUs     uint   `yaml:"cpus" json:"cpus"`
	Cores    uint   `yaml:"cores" json:"cores"`
	Threads  uint   `yaml:"threads" json:"threads"`
	Sockets  uint   `yaml:"sockets" json:"sockets"`
	Memory   uint   `yaml:"memory" json:"memory"`
}

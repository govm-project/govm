package types

//ConfigDriveMetaData stands for cloud images config drive
type ConfigDriveMetaData struct {
	AvailabilityZone string            `json:"availavility_zone"`
	Hostname         string            `json:"hostname"`
	LaunchIndex      string            `json:"launch_index"`
	Name             string            `json:"name"`
	Meta             map[string]string `json:"meta"`
	PublicKeys       map[string]string `json:"public_keys"`
	UUID             string            `json:"uuid"`
}

//NetworkingOptions specifies network details for new VM
type NetworkingOptions struct {
	IP    string   `yaml:"ip"`
	MAC   string   `yaml:"mac"`
	NetID string   `yaml:"net-id"`
	DNS   []string `yaml:"dns"`
}

//VMSize specifies all custome VM size fields
type VMSize struct {
	CPUModel string `yaml:"cpu-model"`
	Sockets  int    `yaml:"sockets"`
	Cpus     int    `yaml:"cpus"`
	Cores    int    `yaml:"cores"`
	Threads  int    `yaml:"threads"`
	RAM      int    `yaml:"ram"`
}

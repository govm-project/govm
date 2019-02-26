package types

// Network represents a network for VMs
type Network struct {
	ID     string   `yaml:"id" json:"id"`
	Subnet string   `yaml:"subnet" json:"subnet"`
	DNS    []string `yaml:"dns" json:"dns"`
}

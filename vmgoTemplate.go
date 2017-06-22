package main

type VmgoTemplate struct {
	Vmgos []Vmgo `yaml:"go-vms"`
}

func NewVmgoTemplate(c *VmgoTemplate) VmgoTemplate {
	var newVmgoTemplate VmgoTemplate
	for _, vmgo := range c.Vmgos {
		newVmgoTemplate.Vmgos = append(newVmgoTemplate.Vmgos,
			NewVmgo(
				vmgo.Name,
				vmgo.ParentImage,
				HostOpts(vmgo.Size),
				vmgo.Cloud,
				vmgo.Efi,
				vmgo.Workdir,
				vmgo.SSHKey,
				vmgo.UserData),
		)
	}
	return newVmgoTemplate
}

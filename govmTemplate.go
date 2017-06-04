package main

type GoVMTemplate struct {
	GoVMs []GoVM `yaml:"go-vms"`
}

func NewGoVMTemplate(c *GoVMTemplate) GoVMTemplate {
	var newGoVMTemplate GoVMTemplate
	for _, govm := range c.GoVMs {
		newGoVMTemplate.GoVMs = append(newGoVMTemplate.GoVMs,
			NewGoVM(
				govm.Name,
				govm.ParentImage,
				HostOpts(govm.Size),
				govm.Cloud,
				govm.Efi,
				govm.Workdir,
				govm.SSHKey,
				govm.UserData),
		)
	}
	return newGoVMTemplate
}

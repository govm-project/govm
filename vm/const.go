package vm

// VM launcher environment variables
const (
	DefaultSSHPublicKeyFile = "id_rsa.pub"
)

// Qemu Command Parameters
const (
	KVMCPUOpts = `KVM_CPU_OPTS=-cpu %s
                      -smp sockets=%v,cpus=%v,cores=%v,threads=%v,maxcpus=%v
                      -m %d`
	CloudInitOpts = `CLOUD_INIT_OPTS=-drive
                         file=/data/seed.iso,if=virtio,format=raw`
)

// Mount binds
const (
	ImageMount    = "%v:/image/image"
	DataMount     = "%v:/data"
	MetadataMount = "%v/%v:/cloud-init/openstack/latest/meta_data.json"
	MedatataFile  = "meta_data.json"
	UserDataMount = "%s:/cloud-init/openstack/latest/user_data"
)

// Default size of qcow2 root disk in GB.
const DiskDefaultSizeGB = 50

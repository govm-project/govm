# Container VM Launcher

``govm`` is a Docker-based tool that will launch your pet VM inside docker containers. It will use Docker networking layer and will map it to your VM.

**Key Features**

- Uses Docker networking
- Cloud Init support
- Copy-on-write images
- EFI support
- Cloud-Init and RAW images support

**Requirements**

- Go 1.11+
- Docker

Build the project
-----------------
```
export GO111MODULE=on
mkdir -p $GOPATH/src/github.com/
git clone https://github.com/govm-project/govm.git
cd govm/
go build -o govm
```


Launch your first VM (Ubuntu 16.04 cloud image)
-----------------------------------------------
```
# Download Ubuntu 16.04 cloud image
curl -Ok https://cloud-images.ubuntu.com/xenial/current/xenial-server-cloudimg-amd64-disk1.img
# Launch your VM
govm create --image xenial-server-cloudimg-amd64-disk1.img --cloud
```


Sub-commands
============

create
------
Creates a new Virtual Machine inside a privileged docker container.

| Flag              | Description                                                     | Required |
|-------------------|-----------------------------------------------------------------|----------|
| --image value     | Path to image                                                   | Yes      |
| --user-data value | Path to user data file                                          | No       |
| --efi             | Use efi bootloader                                              | No       |
| --cloud           | Create config-drive for cloud images                            | No       |
| --flavor value    | VM specs descriptor                                             | Yes      |
| --key value       | SSH key to be included in a cloud image                         | No       |
| --name value      | VM name                                                         | No       |
| --namespace value | VM namespace (this will normally be the user's username)        | No       |
| --cpumodel value  | Model of the virtual cpu. See: ``qemu-system-x86_64 -cpu help`` | No       |
| --sockets value   | Number of sockets. (default: 1)                                 | No       |
| --cpus value      | Number of cpus (default: 1)                                     | No       |
| --cores value     | Number of cores (default: 2)                                    | No       |
| --threads value   | Number of threads (default: 2)                                  | No       |
| --ram value       | Allocated RAM (default: 1024)                                   | No       |
| --debug           | Debug mode                                                      | No       |

remove
------
Removes a the whole privileged docker container and its virtual machine data.

| Flag  | Description                                                                                     | Required |
|-------|-------------------------------------------------------------------------------------------------|----------|
| value | If the value (name of container) is specified, it will remove it. See: ``govm list`` to get name | Yes      |
| --all | Removes all ``govm`` created virtual machines                                                    | No       |

list
----
Lists all virtual machines that were created with the ``govm`` tool. It also shows the VNC access url and name.

| Flag              | Description                       | Required |
|-------------------|-----------------------------------|----------|
| --all             | Show VMs from all namespaces      | No       |
| --namespace value | Show VMs from the given namespace | No       |

*Output example:*
```
ID          IP           VNC_URL                 NAME
bd0088a3eb  172.17.0.2   http://localhost:40669  clear-test
db3ea1be1d  172.17.0.3   http://localhost:35957  cirros-test
f8a4cd7e93  172.17.0.4   http://localhost:42161  test-30310
```

compose
-------
Deploys one or multiple virtual machines with a given compose template file.

| Flag     | Description   | Required |
|----------|---------------|----------|
| -f value | Template file | Yes      |

YAML template file example:
- [2 VMs deployment](data/compose/example_v1.yml)

connect
-------
Connects through ssh to the specified virtual machine.

| Flag         | Description                               | Required |
|--------------|-------------------------------------------|----------|
| --user value | ssh login user                            | Yes      |
| --key value  | private key path (default: ~/.ssh/id_rsa) | No       |

help
----

```
NAME:
   govm - Virtual Machines on top of Docker containers

USAGE:
   govm [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     create, c      Create a new govm
     delete, d      Delete govms
     list, ls       List govms
     compose, co    Deploy Govms from yaml templates
     connect, conn  Get a shell from a Govm
     help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --workdir value  Alternate working directory. Default: ~/govm
   --help, -h       show help
   --version, -v    print the version
```

More cloud init stuff?
----------------------

If you want to boot cloud images, edit the template files under $HOME/govm/cloud-init/openstack/latest/ to fit your own needs and use the `-cloud` flag.
For more information, please see the cloud-init documentation: https://cloudinit.readthedocs.io/en/latest/

based on https://github.com/BBVA/kvm

Questions, issues or suggestions
--------------------------------

http://github.com/govm-project/govm/issues

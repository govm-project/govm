# GoVM Project
[![Build Status](https://semaphoreci.com/api/v1/govmproject/govm/branches/master/badge.svg)](https://semaphoreci.com/govmproject/govm)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgovm-project%2Fgovm.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgovm-project%2Fgovm?ref=badge_shield)

``govm`` is a Docker-based tool that will launch your pet VM inside docker containers. It will use Docker networking layer and will map it to your VM.

**Key Features**

- Uses Docker networking
- Cloud Init support
- Copy-on-write images
- EFI support
- Cloud-Init and RAW images support
- Multi-Instances Compose
- Custom hypervisor-level and qemu flags


**The only Requirement**
- Docker


Quick Install (GNU/Linux based distribution)
-------------------------------------------

1. Download the latest `govm` release from https://github.com/govm-project/govm/releases/latest
2. Install it
```
tar -xzvf govm_<latest_govm_release>_linux_amd64.tar.gz
sudo cp govm /usr/local/bin/

# Make sure you have `/usr/local/bin` in your PATH, if not, run this command
# export PATH=/usr/local/bin/:$PATH
```
3. Done, you're ready to go.


Build your own `govm`
--------------------

**Build Requirements**
- Go 1.11+

```
export GO111MODULE=on
mkdir -p $GOPATH/src/github.com/govm-project/
cd $GOPATH/src/github.com/govm-project/
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
Removes the whole privileged docker container and its virtual machine data.

| Flag  | Description                                                                                     | Required |
|-------|-------------------------------------------------------------------------------------------------|----------|
| value | If the value (name of container) is specified, it will remove it. See: ``govm list`` to get name | Yes      |
| --all | Removes all ``govm`` created virtual machines                                                    | No       |

start
-----
Starts a stopped GoVM Instance

| Flag  | Description                 | Required |
|-------|-----------------------------|----------|
| value | GoVM instance's name or ID  | Yes      |

list
----
Lists all virtual machines that were created with the ``govm`` tool. It also shows the VNC access url and name.

| Flag                     | Description                                    | Required |
|--------------------------|------------------------------------------------|----------|
| --all                    | Show VMs from all namespaces                   | No       |
| --namespace value        | Show VMs from the given namespace              | No       |
| --format value, -f value | String containing the template code to execute | No       |

*Output example*
```
# govm list
ID         Name                   Namespace IP
b9b5d3a288 test-14731             onmunoz   172.17.0.6
4d6731b571 test-29652             onmunoz   172.17.0.5
ef004385d4 test-20024             onmunoz   172.17.0.4
5e5f42047b happy-poitras          onmunoz   172.17.0.3
e90608db45 wonderful-varahamihira onmunoz   172.17.0.2
```

*Filtered output*
```
# govm list -f '{{select (filterRegexp . "Name" "test-*") "IP"}}'
172.17.0.6
172.17.0.5
172.17.0.4
```

compose
-------
Deploys one or multiple virtual machines with a given compose template file.

| Flag     | Description   | Required |
|----------|---------------|----------|
| -f value | Template file | Yes      |

YAML template file example:
- [2 VMs deployment](data/compose/example_v1.yml)

ssh
---

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
     create, c                Create a new VM
     list, ls                 List VMs
     remove, delete, rm, del  Remove VMs
     start, up, s             Start a GoVM Instance
     compose, co              Deploy VMs from a compose config file
     ssh                      ssh into a running VM
     help, h                  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --workdir value  Alternate working directory. Default: ~/govm
   --help, -h       show help
   --version, -v    print the version
```

More cloud init stuff?
----------------------

If you want to boot cloud images, edit the template files under $HOME/govm/cloud-init/openstack/latest/ to fit your own needs and use the `--cloud` flag.
For more information, please see the cloud-init documentation: https://cloudinit.readthedocs.io/en/latest/

based on https://github.com/BBVA/kvm

Questions, issues or suggestions
--------------------------------

http://github.com/govm-project/govm/issues


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgovm-project%2Fgovm.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgovm-project%2Fgovm?ref=badge_large)

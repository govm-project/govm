# govm

Introduction
------------
``govm`` is a Docker-based tool that will launch your pet VM inside docker containers. It will use Docker networking layer and will map it to your VM.

Key Features
------------
- Uses Docker networking
- Cloud Init support
- Copy-on-write images
- EFI support

Requirements
---------------
- Go 1.7+
- Docker
- (websockify)[https://github.com/novnc/websockify]

Get the project
---------------
```
mkdir -p $GOPATH/src/clrgitlab.amr.corp.intel.com/clr-cloud
cd $GOPATH/src/clrgitlab.amr.corp.intel.com/clr-cloud
git clone http://onmunoz@clrgitlab.amr.corp.intel.com/clr-cloud/govm.git
go get -v clrgitlab.amr.corp.intel.com/clr-cloud/govm
```

Setup ``govm`` data directories
----------------------------------
```
$GOPATH/src/gitlab.com/verbacious/govm/setup.sh
```

Launch your first VM (Ubuntu 16.04 cloud image)
-----------------------------------------------
```
# Download Ubuntu 16.04 cloud image
curl -Ok https://cloud-images.ubuntu.com/xenial/20170303.1/xenial-server-cloudimg-amd64-disk1.img
# Launch your VM
govm --image ./xenial-server-cloudimg-amd64-disk1.img --cloud my-test
```

**Output**
```
// TODO: Update output
```

# Log into your vm
```
ssh cloud@172.17.0.2
```

``govm`` help
-------------

```
> govm --help
NAME:
   govm - Virtual Machines on top of Docker containers

USAGE:
   govm [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     create, c  Create a new govm
     delete, d  Delete govms
     list, ls   List govms
     help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --workdir value  Alternate working directory. Default: ~/govm
   --help, -h       show help
   --version, -v    print the version
```

More cloud init stuff?
----------------------

If you want to boot cloud images, edit the template files under ./cloud-init/openstack/latest/ to fit your own needs and use the `-cloud` flag.
For more information, please see the cloud-init documentation: https://cloudinit.readthedocs.io/en/latest/

based on https://github.com/BBVA/kvm

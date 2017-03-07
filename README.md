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

Getting Started
---------------
-  Make sure you have the latest golang binaries(at least 1.7.4+).
- Setup your golang environment. https://golang.org/doc/code.html#GOPATH

Get the project
---------------
```
go get -v -u github.com/obedmr/govm
```

Setup ``govm`` data directories
----------------------------------
```
$GOPATH/src/github.com/obedmr/govm/setup.sh
```

Launch your first VM (Ubuntu 16.04 cloud image)
-----------------------------------------------
```
# Download Ubuntu 16.04 cloud image
curl -Ok https://cloud-images.ubuntu.com/xenial/20170303.1/xenial-server-cloudimg-amd64-disk1.img
# Launch your VM
sudo -E govm --image ./xenial-server-cloudimg-amd64-disk1.img --name ubuntu16.04 -cloud -v
```

**Output**
```
[create -f qcow2 -o backing_file=/tmp/xenial-server-cloudimg-amd64-disk1.img temp.img]
[run --name ubuntu16.04 -td --privileged -v /tmp/xenial-server-cloudimg-amd64-disk1.img:/tmp/xenial-server-cloudimg-amd64-disk1.img -v /var/lib/govm/data/ubuntu16.04/40a78af6-dae0-4d07-be96-d659f4a54752/ubuntu16.04.img:/image/image -e AUTO_ATTACH=yes -v /var/lib/govm/data/ubuntu16.04/40a78af6-dae0-4d07-be96-d659f4a54752:/var/lib/govm/data/ubuntu16.04/40a78af6-dae0-4d07-be96-d659f4a54752 -v /var/lib/govm/data/ubuntu16.04/40a78af6-dae0-4d07-be96-d659f4a54752/cidata.iso:/cidata.iso obedmr/govm -vnc unix:/var/lib/govm/data/ubuntu16.04/40a78af6-dae0-4d07-be96-d659f4a54752/vnc -drive file=/cidata.iso,if=virtio --enable-kvm -m 4096 -smp cpus=4,cores=2,threads=2 -cpu host]
[ubuntu16.04] info:
IP 172.17.0.2
```

# Log into your vm
```
ssh cloud@172.17.0.2
```

``govm`` help
-------------

```
> govm --help
Usage of govm:
  -cloud
        Cloud VM (Optional)
  -efi
        EFI-enabled vm (Optional)
  -image string
        qcow2 image file path (default "image.qcow2")
  -large
        Small VM flavor (8G ram, cpus=8,cores=4,threads=4)
  -name string
        VM's name
  -resize int
        Resize value in GB (Only for QCOW Images).
  -small
        Small VM flavor (2G ram, cpus=4,cores=2,threads=2)
  -tiny
        Tiny VM flavor (512MB ram, cpus=1,cores=1,threads=1)
  -v    Enable verbosity
```

More cloud init stuff?
----------------------

If you want to boot cloud images, edit the template files under ./cloud-init/openstack/latest/ to fit your own needs and use the `-cloud` flag.
For more information, please see the cloud-init documentation: https://cloudinit.readthedocs.io/en/latest/

based on https://github.com/BBVA/kvm

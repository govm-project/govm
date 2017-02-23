# govm


Introduction
------------
``govm`` launches vms inside docker containers

Getting Started
---------------
/!\ Make sure you have the latest golang binaries(at least 1.7.4+).

Get the project:
```
go get -v -u github.com/obedmr/govm
```
Launch your first VM:
```
./govm -name myvm -image=/path/to/image
```
/!\ See ./govm -help for more arguments

/!\ If you want to boot cloud images, edit the template files under ./cloud-init/openstack/latest/ to fit your own needs and use the `-cloud` flag.
For more information, please see the cloud-init documentation: https://cloudinit.readthedocs.io/en/latest/

based on https://github.com/BBVA/kvm

#!/bin/bash

# Environment checks
if [ -z $GOPATH ]; then
	echo "GOPATH is not set"
	exit 1
fi

## Make sure GOBIN is added in PATH ##
is_gobin_in_path=$(echo $PATH | grep $GOPATH/bin)
if [ -z $is_gobin_in_path ]; then
	echo "GOBIN is not added in PATH."
	echo "Try:"
	echo "echo \"export GOBIN=\$GOPATH/bin\" >> ~/.bashrc"
	echo "echo \"export PATH=\$PATH:\$GOBIN\" >> ~/.bashrc"
	exit 1
fi

# Setup
go get -v -u github.com/obedmr/govm
sudo mkdir -p /var/lib/govm/data
sudo mkdir -p /var/lib/govm/images
sudo cp -r $GOPATH/src/github.com/obedmr/govm/cloud-init/ /var/lib/govm

# Add public key to cloud-init files
publicKey=$(cat ~/.ssh/id_rsa.pub)
sudo sed -i 's|YOUR-PUBLIC-KEY-GOES-HERE|'"$publicKey"'|g' /var/lib/govm/cloud-init/openstack/latest/user_data
sudo sed -i 's|YOUR-PUBLIC-KEY-GOES-HERE|'"$publicKey"'|g' /var/lib/govm/cloud-init/openstack/latest/meta_data.json

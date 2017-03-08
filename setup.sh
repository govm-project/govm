#!/bin/bash

workdir=~/govm
publicKey=$(cat ~/.ssh/id_rsa.pub)

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
mkdir -p $workdir/data
mkdir $workdir/images
cp -r $GOPATH/src/github.com/verbacious/govm/cloud-init/ $workdir

# Add public key to cloud-init files
sed -i 's|YOUR-PUBLIC-KEY-GOES-HERE|'"$publicKey"'|g' $workdir/cloud-init/openstack/latest/user_data
sed -i 's|YOUR-PUBLIC-KEY-GOES-HERE|'"$publicKey"'|g' $workdir/cloud-init/openstack/latest/meta_data.json

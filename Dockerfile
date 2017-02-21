FROM bbvainnotech/kvm:latest
MAINTAINER obed.n.munoz@gmail.com

RUN curl -O https://download.clearlinux.org/image/OVMF.fd -o /image/OVMF.fd


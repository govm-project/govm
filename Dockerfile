FROM alpine
LABEL govm.image.authors="obed.n.munoz@gmail.com, erick.cardona.ruiz@gmail.com"
ENV container=docker

RUN apk update \
&& apk add qemu-system-x86_64 cdrkit dnsmasq net-tools bridge-utils \
iproute2 curl bash qemu-img ovmf \
&& ( apk add qemu-hw-display-qxl || true )


COPY startvm /usr/local/bin/startvm
RUN chmod u+x /usr/local/bin/startvm

VOLUME /image

ENTRYPOINT ["/usr/local/bin/startvm"]

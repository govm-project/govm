FROM alpine:3.5
MAINTAINER obed.n.munoz@gmail.com, erick.cardona.ruiz@gmail.com

RUN apk update
RUN apk add git python3 bash
RUN ln -s /usr/bin/python3 /bin/python

RUN git clone https://github.com/novnc/noVNC.git
RUN git clone https://github.com/novnc/websockify.git /noVNC/utils/websockify
RUN sed -i -- "s/ps -p/ps -o pid | grep/g" /noVNC/utils/launch.sh
RUN cp  /noVNC/vnc.html /noVNC/index.html

ENTRYPOINT ["./noVNC/utils/launch.sh"]

FROM ubuntu:18.04

# Install Go
ENV GOLANG_VERSION 1.18.3
RUN apt update && apt install -y build-essential wget
RUN wget https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz
RUN rm -f go${GOLANG_VERSION}.linux-amd64.tar.gz
ENV PATH "$PATH:/usr/local/go/bin"

# ACA
WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-agent/plugins/alcor-control-agent/build/bin/AlcorControlAgent /merak-bin/AlcorControlAgent
COPY services/merak-agent/plugins/alcor-control-agent/build/aca-machine-init.sh /aca-machine-init.sh
RUN apt install -y git make bash gcc libc-dev openvswitch-switch=2.9.8-0ubuntu0.18.04.2 libevent-dev rsyslog
RUN ./aca-machine-init.sh

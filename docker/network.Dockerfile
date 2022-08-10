FROM golang:1.18-alpine

WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-network/merak-network /merak-bin/merak-network
CMD [ "/merak-bin/merak-network" ]
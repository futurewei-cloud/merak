FROM golang:1.18-alpine

# Merak
WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-topo/build/merak-topo /merak-bin/merak-topo
CMD [ "/merak-bin/merak-topo" ]

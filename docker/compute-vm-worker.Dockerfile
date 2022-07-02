FROM golang:1.18-alpine

# Merak
WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-compute/build/merak-compute-vm-worker /merak-bin/merak-compute-vm-worker
CMD [ "/merak-bin/merak-compute-vm-worker" ]
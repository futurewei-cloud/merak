FROM golang:1.18-alpine

# Merak
WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-compute/build/merak-compute /merak-bin/merak-compute
CMD [ "/merak-bin/merak-compute" ]
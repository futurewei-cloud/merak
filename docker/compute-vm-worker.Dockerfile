FROM golang:1.18-alpine

COPY . /merak
WORKDIR /merak
RUN go mod download
RUN apk add --no-cache git make bash gcc libc-dev
RUN make
CMD [ "/merak/services/merak-compute/build/merak-compute-vm-worker" ]

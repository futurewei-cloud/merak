FROM golang:1.18-alpine

# Merak
COPY . /merak
WORKDIR /merak
RUN go mod download
RUN apk add --no-cache git make bash gcc libc-dev
ENTRYPOINT [ "tail" ]
CMD [ "-f", "/dev/null" ]
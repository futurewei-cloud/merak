module := merak-compute

merak.compute:: build

build:
	go build -o services/merak-compute/build/merak-compute services/merak-compute/main.go

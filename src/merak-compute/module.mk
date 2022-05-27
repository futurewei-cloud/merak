module := merak-compute

merak.compute:: build

build:
	go build -o src/merak-compute/build/merak-compute src/merak-compute/main.go

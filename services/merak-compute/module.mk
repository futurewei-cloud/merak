module := merak-compute

merak.compute:: main vm-worker

main:
	go build -o services/merak-compute/build/merak-compute services/merak-compute/compute.go
vm-worker:
	go build -o services/merak-compute/build/merak-compute-vm-worker services/merak-compute/workers/vm/worker.go

module := merak-compute

merak-compute: compute vm-worker

compute:
	go build -o services/merak-compute/build/merak-compute services/merak-compute/compute.go
vm-worker:
	go build -o services/merak-compute/build/merak-compute-vm-worker services/merak-compute/workers/worker.go

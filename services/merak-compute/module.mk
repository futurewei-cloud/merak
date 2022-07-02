module := merak-compute

merak-compute: compute vm-worker

compute:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/merak-compute/build/merak-compute services/merak-compute/compute.go
vm-worker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/merak-compute/build/merak-compute-vm-worker services/merak-compute/workers/worker.go

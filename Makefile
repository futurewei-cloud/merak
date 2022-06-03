modules := services
-include $(patsubst %, %/module.mk, $(modules))

all:: merak-compute proto

proto:
	protoc --go_out=api/proto/v1/ --go-grpc_out=api/proto/v1/ -I api/proto/v1/ api/proto/v1/*.proto

deploy-dev:
	kubectl apply -f deployments/kubernetes/compute.dev.yaml

docker-build:
# Compute
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t phudtran/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push phudtran/merak-compute:dev
	docker push phudtran/merak-compute-vm-worker:dev
# Netowrk
	docker build -t yanmo96/merak-network:dev -f docker/netowrk.Dockerfile .
	docker push yanmo96/merak-network:dev

test:
	go test -v services/merak-compute/tests/compute_test.go

clean:
	rm api/proto/v1/merak/*.pb.go
	rm services/merak-compute/build/*

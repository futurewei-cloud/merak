modules := services
-include $(patsubst %, %/module.mk, $(modules))

all:: services proto

proto:
	protoc --go_out=api/proto/v1/compute \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/compute \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/compute.proto

	protoc --go_out=api/proto/v1/common \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/common \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/common.proto

	protoc --go_out=api/proto/v1/network \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/network \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/network.proto

	protoc --go_out=api/proto/v1/topology \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/topology \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/topology.proto

	protoc --go_out=api/proto/v1/agent \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/agent \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/agent.proto

deploy-dev:
	kubectl apply -f deployments/kubernetes/compute.dev.yaml
	kubectl apply -f deployments/kubernetes/scenario.dev.yaml

docker-build:
# Scenario-Manager
	docker build -t cjchung4849/scenario-manager:dev -f docker/scenario.Dockerfile .
	docker push cjchung4849/scenario-manager:dev
# Compute
	make proto
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t phudtran/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push phudtran/merak-compute:dev
	docker push phudtran/merak-compute-vm-worker:dev
# Agent
	docker build -t phudtran/merak-agent:dev -f docker/agent.Dockerfile .
	docker push phudtran/merak-agent:dev

docker-compute:
	make proto
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t phudtran/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push phudtran/merak-compute:dev
	docker push phudtran/merak-compute-vm-worker:dev

docker-agent:
	make proto
	docker build -t phudtran/merak-agent:dev -f docker/agent.Dockerfile .
	docker push phudtran/merak-agent:dev

test:
	go test -v services/merak-compute/tests/compute_test.go

clean:
	rm -rf api/proto/v1/*.pb.go
	rm -rf api/proto/v1/common/*.pb.go
	rm -rf api/proto/v1/agent/*.pb.go
	rm -rf api/proto/v1/compute/*.pb.go
	rm -rf services/merak-compute/build/*
	rm -rf services/merak-agent/build/*
	rm -rf services/scenario-manager/build/*


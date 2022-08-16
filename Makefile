# MIT License
# Copyright(c) 2022 Futurewei Cloud
#     Permission is hereby granted,
#     free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
#     including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
#     to whom the Software is furnished to do so, subject to the following conditions:
#     The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
#     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
#     WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

modules := services
-include $(patsubst %, %/module.mk, $(modules))

.DEFAULT_GOAL := all

GFLAGS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64

.PHONY: all
all:: proto services

proto:
	mkdir -p api/proto/v1/compute
	protoc --go_out=api/proto/v1/compute \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/compute \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/compute.proto

	mkdir -p api/proto/v1/common
	protoc --go_out=api/proto/v1/common \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/common \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/common.proto

	mkdir -p api/proto/v1/network
	protoc --go_out=api/proto/v1/network \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/network \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/network.proto

	mkdir -p api/proto/v1/topology
	protoc --go_out=api/proto/v1/topology \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/topology \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/topology.proto

	mkdir -p api/proto/v1/agent
	protoc --go_out=api/proto/v1/agent \
	--go_opt=paths=source_relative \
	--go-grpc_out=api/proto/v1/agent \
	--go-grpc_opt=paths=source_relative \
	-I api/proto/v1/ api/proto/v1/agent.proto

.PHONY: deploy-dev
deploy-dev:
	kubectl apply -f deployments/kubernetes/scenario.dev.yaml
	kubectl apply -f deployments/kubernetes/compute.dev.yaml
	kubectl apply -f deployments/kubernetes/network.dev.yaml
	kubectl apply -f deployments/kubernetes/topo.dev.yaml

.PHONY: docker-scenario
docker-scenario:
# Scenario-Manager
	make proto
	make scenario
	docker build -t meraksim/scenario-manager:dev -f docker/scenario.Dockerfile .
	docker push meraksim/scenario-manager:dev


.PHONY: docker-compute
docker-compute:
	make proto
	make compute
	docker build -t meraksim/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t meraksim/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push meraksim/merak-compute:dev
	docker push meraksim/merak-compute-vm-worker:dev


.PHONY: docker-agent
docker-agent:
	make proto
	make agent
	docker build -t meraksim/merak-agent:dev -f docker/agent.Dockerfile .
	docker push meraksim/merak-agent:dev


.PHONY: docker-network
docker-network:
	make proto
	make network
	docker build -t meraksim/merak-network:dev -f docker/network.Dockerfile .
	docker push meraksim/merak-network:dev


.PHONY: docker-topo
docker-topo:
	make proto
	make topo
	docker build -t meraksim/merak-topo:dev -f docker/topo.Dockerfile .
	docker push meraksim/merak-topo:dev


.PHONY: docker-all
docker-all:
	make
	docker build -t meraksim/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t meraksim/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker build -t meraksim/merak-agent:dev -f docker/agent.Dockerfile .
	docker build -t meraksim/merak-topo:dev -f docker/topo.Dockerfile .
	docker build -t meraksim/merak-network:dev -f docker/network.Dockerfile .
	docker build -t meraksim/scenario-manager:dev -f docker/scenario.Dockerfile .
	docker push meraksim/merak-compute:dev
	docker push meraksim/merak-compute-vm-worker:dev
	docker push meraksim/merak-agent:dev
	docker push meraksim/scenario-manager:dev
	docker push meraksim/merak-network:dev
	docker push meraksim/merak-topo:dev

.PHONY: docker-all-ci
docker-all-ci:
	make
	docker build -t meraksim/merak-compute:ci -f docker/compute.Dockerfile .
	docker build -t meraksim/merak-compute-vm-worker:ci -f docker/compute-vm-worker.Dockerfile .
	docker build -t meraksim/merak-topo:ci -f docker/topo.Dockerfile .
	docker build -t meraksim/merak-network:ci -f docker/network.Dockerfile .
	docker build -t meraksim/scenario-manager:ci -f docker/scenario.Dockerfile .
	docker build -t meraksim/merak-agent:ci -f docker/agent.Dockerfile .
	docker push meraksim/merak-agent:ci
	docker push meraksim/merak-compute:ci
	docker push meraksim/merak-compute-vm-worker:ci
	docker push meraksim/merak-network:ci
	docker push meraksim/merak-topo:ci
	docker push meraksim/scenario-manager:ci



.PHONY: clean
clean:
	rm -rf api/proto/v1/merak/*.pb.go
	rm -rf services/merak-compute/build/*
	rm -rf services/merak-agent/build/*
	rm -rf services/scenario-manager/build/*
	rm -rf services/merak-network/build/*
	rm -rf services/merak-topo/build/*

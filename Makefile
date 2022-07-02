modules := services
-include $(patsubst %, %/module.mk, $(modules))

all:: services proto

proto:
	protoc --go_out=api/proto/v1/ --go-grpc_out=api/proto/v1/ -I api/proto/v1/ api/proto/v1/*.proto

deploy-dev:
	kubectl apply -f deployments/kubernetes/compute.dev.yaml
	kubectl apply -f deployments/kubernetes/scenario.dev.yaml

<<<<<<< HEAD
docker-build:
# Scenario-Manager
	docker build -t cjchung4849/scenario-manager:dev -f docker/scenario.Dockerfile .
	docker push cjchung4849/scenario-manager:dev
=======
docker-all:
>>>>>>> Initial agent
# Compute
	make all
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t phudtran/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push phudtran/merak-compute:dev
	docker push phudtran/merak-compute-vm-worker:dev
# Agent
	docker build -t phudtran/merak-agent:dev -f docker/agent.Dockerfile .
	docker push phudtran/merak-agent:dev

docker-compute:
	make all
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker build -t phudtran/merak-compute-vm-worker:dev -f docker/compute-vm-worker.Dockerfile .
	docker push phudtran/merak-compute:dev
	docker push phudtran/merak-compute-vm-worker:dev

docker-agent:
	make all
	docker build -t phudtran/merak-agent:dev -f docker/agent.Dockerfile .
	docker push phudtran/merak-agent:dev

test:
	go test -v services/merak-compute/tests/compute_test.go

clean:
	rm api/proto/v1/merak/*.pb.go
	rm services/merak-compute/build/*
	rm services/merak-agent/build/*
	rm services/scenario-manager/build/*

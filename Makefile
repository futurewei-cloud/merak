
# MIT License
# Copyright(c) 2022 Futurewei Cloud
#    Permission is hereby granted,
#    free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
#    including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
#    to whom the Software is furnished to do so, subject to the following conditions:
#    The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
#    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
#    WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

modules := services
-include $(patsubst %, %/module.mk, $(modules))

all:: services proto

proto:
	protoc --go_out=api/proto/v1/ --go-grpc_out=api/proto/v1/ -I api/proto/v1/ api/proto/v1/*.proto

deploy-dev:
	kubectl apply -f deployments/kubernetes/scenario.dev.yaml

<<<<<<< HEAD
docker-build:
# Scenario-Manager
	make scenario
	docker build -t cjchung4849/scenario-manager:dev -f docker/scenario.Dockerfile .
	docker push cjchung4849/scenario-manager:dev
=======
docker-all:
>>>>>>> Initial agent
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
	rm api/proto/v1/merak/*.pb.go
	rm services/merak-compute/build/*
	rm services/scenario-manager/build/*

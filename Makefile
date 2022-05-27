modules := services
-include $(patsubst %, %/module.mk, $(modules))

all:: merak-compute proto

proto:
	protoc --go_out=api/proto/v1/ --go-grpc_out=api/proto/v1/ -I api/proto/v1/ api/proto/v1/*.proto

deploy-dev:
	kubectl apply -f deployments/kubernetes/compute.dev.yaml

docker-build:
	docker build -t phudtran/merak-compute:dev -f docker/compute.Dockerfile .
	docker push phudtran/merak-compute:dev

clean:
	rm api/proto/v1/merak/*.pb.go
	rm services/merak-compute/build/*

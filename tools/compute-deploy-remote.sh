#!/bin/bash

make docker-build

ssh devbox2 "kubectl delete namespace merak"
ssh devbox2 "kubectl delete pod grpc"
sleep 10
ssh devbox2 "sudo crictl rmi --prune"
ssh devbox2 "kubectl apply -f /home/ubuntu/merak/deployments/kubernetes/compute.dev.yaml"
ssh devbox2 "kubectl run grpc --image=phudtran/merak-compute-vm-worker:dev"

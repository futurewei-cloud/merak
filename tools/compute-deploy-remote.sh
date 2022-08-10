#!/bin/bash
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
ssh devbox2 "kubectl delete namespace merak"
ssh devbox2 "kubectl delete pod grpc"
sleep 20
ssh devbox2 "sudo crictl rmi --prune"
ssh devbox2 "sudo crictl images"
ssh devbox2 "kubectl apply -f /home/ubuntu/merak/deployments/kubernetes/compute.dev.yaml"
ssh devbox2 "kubectl run grpc --image=phudtran/merak-compute-vm-worker:dev"
ssh devbox2 "kubectl apply -f /home/ubuntu/merak/examples/poc-demo/merak.node.yaml"
ssh devbox2 "kubectl exec grpc -- bash -c 'go test -v /merak/services/merak-compute/tests/grpc_client_test.go'"

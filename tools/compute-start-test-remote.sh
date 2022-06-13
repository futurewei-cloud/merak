#!/bin/bash

ssh devbox2 "kubectl exec grpc -- bash -c 'go test -v /merak/services/merak-compute/tests/grpc_client_test.go'"
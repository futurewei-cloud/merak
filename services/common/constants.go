/*
MIT License
Copyright(c) 2022 Futurewei Cloud

	Permission is hereby granted,
	free of charge, to any person obtaining a copy of this software and associated documentation files(the "Software"), to deal in the Software without restriction,
	including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and / or sell copies of the Software, and to permit persons
	to whom the Software is furnished to do so, subject to the following conditions:
	The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package constants

const (
	TEMPORAL_ADDRESS                  = "temporal-frontend.default.svc.cluster.local"
	TEMPORAL_PORT                     = 7233
	TEMPORAL_ENV                      = "TEMPORAL"
	TEMPORAL_RPS_ENV                  = "RPS"
	TEMPORAL_CONCURRENCY_ENV          = "CONCURRENCY"
	TEMPORAL_RPS_UPPER_ENV            = "RPS_UPPER"
	TEMPORAL_RPS_LOWER_ENV            = "RPS_LOWER"
	TEMPORAL_CONCURRENCY_UPPER_ENV    = "CONCURRENCY_UPPER"
	TEMPORAL_CONCURRENCY_LOWER_ENV    = "CONCURRENCY_LOWER"
	TEMPORAL_CONCURRENT_WORKFLOWS_ENV = "CONCURRENT_WORKFLOWS"
	LOCALHOST                         = "127.0.0.1"
	TEMPORAL_NAMESPACE                = "merak"
	TEMPORAL_TQ_ENV                   = "TQ"

	KUBE_NODE_LABEL_KEY = "DEDICATED"
	KUBE_NODE_LABEL_VAL = "MERAK"

	COMPUTE_GRPC_SERVER_ADDRESS = "merak-compute-service.merak.svc.cluster.local"
	TOPLOGY_GRPC_SERVER_PORT    = 40052
	TOPLOGY_GRPC_SERVER_ADDRESS = "merak-topology-service.merak.svc.cluster.local"
	NETWORK_GRPC_SERVER_PORT    = 40053
	NETWORK_GRPC_SERVER_ADDRESS = "merak-network-service.merak.svc.cluster.local"
	AGENT_GRPC_SERVER_PORT      = 40054
	COMPUTE_GRPC_SERVER_PORT    = 40051
	PROMETHEUS_PORT             = 9001

	COMPUTE_REDIS_ADDRESS = "compute-redis-main.merak.svc.cluster.local"
	COMPUTE_REDIS_PORT    = 30051

	COMPUTE_REDIS_NODE_IP_SET = "NodeIPSet"
	COMPUTE_REDIS_VM_SET      = "VMSet"

	ALCOR_PORT_MANAGER_PORT        = 30006
	ALCOR_PORT_ID_SUBSTRING_LENGTH = 11

	HTTP_OK             = 200
	HTTP_CREATE_SUCCESS = 201

	MIN_PORT = 1
	MAX_PORT = 65535

	GRPC_MAX_RECV_MSG_SIZE = 1024 * 1024 * 500
	GRPC_MAX_SEND_MSG_SIZE = 1024 * 1024 * 500

	LOG_LEVEL_ENV     = "LOG_LEVEL"
	LOG_LEVEL_DEFAULT = "INFO"
	LOG_LOCATION      = "/var/log/merak/"

	MODE_ENV        = "MODE"
	MODE_STANDALONE = "STANDALONE"
	MODE_ALCOR      = "ALCOR"

	AGENT_STANDALONE_IP        = "10.0.0.2"
	AGENT_STANDALONE_MAC       = "aa:bb:cc:dd:ee:ff"
	AGENT_STANDALONE_REMOTE_ID = "NO ALCOR"
	AGENT_STANDALONE_GW        = "10.0.0.1"
	AGENT_STANDALONE_CIDR      = "10.0.0.0/8"

	WORKER_IMAGE_ENV                    = "WORKER_IMAGE"
	WORKER_DEFAULT_IMAGE                = "meraksim/merak-compute-vm-worker:dev"
	WORKER_DEFAULT_RPS                  = "100000"
	WORKER_DEFAULT_CONCURRENCY          = "1000"
	WORKER_DEFAULT_CONCURRENT_WORKFLOWS = "1000"
	WORKER_POD_PREFIX                   = "compute-worker-"

	POD_PULL_POLICY_ALWAYS = "Always"
)

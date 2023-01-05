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
package common

import (
	"time"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	"github.com/go-redis/redis/v9"
)

var RedisClient redis.Client
var ClientMapGRPC map[string]agent_pb.MerakAgentServiceClient

const (
	TEMPORAL_SUCESS_CODE = "SUCCESS"
	TEMPORAL_FAIL_CODE   = "FAILED"

	TEMPORAL_WF_TASK_TIMEOUT = time.Minute * 2000
	TEMPORAL_WF_EXEC_TIMEOUT = time.Minute * 2000
	TEMPORAL_WF_RUN_TIMEOUT  = time.Minute * 2000

	TEMPORAL_WF_RETRY_INTERVAL = time.Second
	TEMPORAL_WF_BACKOFF        = 1
	TEMPORAL_WF_MAX_INTERVAL   = time.Second * 10
	TEMPORAL_WF_MAX_ATTEMPT    = 1

	TEMPORAL_ACTIVITY_RETRY_INTERVAL = time.Second
	TEMPORAL_ACTIVITY_BACKOFF        = 1
	TEMPORAL_ACTIVITY_MAX_INTERVAL   = time.Second * 10
	TEMPORAL_ACTIVITY_MAX_ATTEMPT    = 1
	TEMPORAL_ACTIVITY_TIMEOUT        = 2400 * time.Minute

	VM_TASK_QUEUE           = "COMPUTE_TASK_QUEUE"
	VM_INFO_WORKFLOW_ID     = "VM_INFO_WORKFLOW"
	VM_CREATE_WORKFLOW_ID   = "VM_CREATE_WORKFLOW"
	VM_GENERATE_WORKFLOW_ID = "VM_GENERATE_WORKFLOW"
	VM_UPDATE_WORKFLOW_ID   = "VM_UPDATE_WORKFLOW"
	VM_DELETE_WORKFLOW_ID   = "VM_DELETE_WORKFLOW"

	COMPUTE_REDIS_POOL_SIZE    = 10000
	COMPUTE_REDIS_POOL_TIMEOUT = time.Second * 60

	DEFAULT_WORKER_CONCURRENCY = "1000"
	DEFAULT_WORKER_RPS         = "1000"
)

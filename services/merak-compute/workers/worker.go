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

package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/delete"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var ctx = context.Background()

func main() {
	common.ClientMapGRPC = make(map[string]agent_pb.MerakAgentServiceClient)
	temporal_address, ok := os.LookupEnv(constants.TEMPORAL_ENV)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		temporal_address = constants.LOCALHOST
	}
	rps, ok := os.LookupEnv(constants.TEMPORAL_CONCURRENCY_ENV)
	if !ok {
		log.Println("RPS environment variable not set, using default.")
		rps = common.DEFAULT_WORKER_RPS
	}
	rps_int, err := strconv.ParseFloat(rps, 64)
	if err != nil {
		log.Fatalln("RPS " + rps + " is NaN!")
	}
	concurrency, ok := os.LookupEnv(constants.TEMPORAL_CONCURRENCY_ENV)
	if !ok {
		log.Println("Concurrency environment variable not set, using default.")
		concurrency = common.DEFAULT_WORKER_CONCURRENCY
	}
	concurrency_int, err := strconv.Atoi(concurrency)
	if err != nil {
		log.Fatalln("Concurrency " + concurrency + " is NaN!")
	}

	log.Println("Starting worker with " +
		rps + " activities/sec and max " +
		concurrency + " concurrent activities")
	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(constants.TEMPORAL_PORT))

	c, err := client.Dial(client.Options{
		HostPort:  sb.String(),
		Namespace: constants.TEMPORAL_NAMESPACE,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	log.Println("Connected to Temporal namespace " + constants.TEMPORAL_NAMESPACE)
	defer c.Close()

	//Connect to Redis
	var redisAddress strings.Builder
	redisAddress.WriteString(constants.COMPUTE_REDIS_ADDRESS)
	redisAddress.WriteString(":")
	redisAddress.WriteString(strconv.Itoa(constants.COMPUTE_REDIS_PORT))

	common.RedisClient = *redis.NewClient(&redis.Options{
		Addr:     redisAddress.String(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err = common.RedisClient.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalln("ERROR: Unable to create Redis client", err)
	}
	defer common.RedisClient.Close()
	log.Println("Connected to DB!")

	w := worker.New(c, common.VM_TASK_QUEUE, worker.Options{
		MaxConcurrentActivityExecutionSize:      concurrency_int,
		WorkerActivitiesPerSecond:               rps_int,
		MaxConcurrentLocalActivityExecutionSize: concurrency_int,
		WorkerLocalActivitiesPerSecond:          rps_int,
	})
	w.RegisterWorkflow(create.Create)
	w.RegisterWorkflow(delete.Delete)
	w.RegisterActivity(activities.VmCreate)
	w.RegisterActivity(activities.VmDelete)
	log.Println("Registered VM Workflows and activities.")
	log.Println("Starting VM Worker.")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

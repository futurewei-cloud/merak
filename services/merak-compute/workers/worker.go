package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/activities"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/create"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/delete"
	"github.com/futurewei-cloud/merak/services/merak-compute/workflows/info"
	"github.com/go-redis/redis/v9"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var ctx = context.Background()

func main() {

	temporal_address, ok := os.LookupEnv(constants.TEMPORAL_ENV)
	if !ok {
		log.Println("Temporal environment variable not set, using default address.")
		temporal_address = constants.TEMPORAL_ADDRESS
	}
	var sb strings.Builder
	sb.WriteString(temporal_address)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(constants.TEMPORAL_PORT))

	c, err := client.NewClient(client.Options{
		HostPort: sb.String(),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
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

	w := worker.New(c, common.VM_TASK_QUEUE, worker.Options{})
	w.RegisterWorkflow(create.Create)
	w.RegisterWorkflow(info.Info)
	w.RegisterWorkflow(delete.Delete)
	w.RegisterActivity(activities.VmCreate)
	w.RegisterActivity(activities.VmInfo)
	w.RegisterActivity(activities.VmDelete)
	log.Println("Registered VM Workflows and activities.")
	log.Println("Starting VM Worker.")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

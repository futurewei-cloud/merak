package activities

import (
	"context"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
	"log"
	"strings"
	"sync"
)

func DoServices(ctx context.Context, services []*pb.InternalServiceInfo, wg *sync.WaitGroup) (string, error) {
	log.Println("DoServices")
	//defer wg.Done()
	log.Printf("DoServices %s", services)
	idServiceMap := make(map[string]*pb.InternalServiceInfo)
	runSequenceMap := make(map[string]string)
	var runId string
	numberOfService := 0
	for _, service := range services {
		if service.WhereToRun == "network" {
			idServiceMap[service.Id] = service
			if service.WhenToRun != "INIT" {
				log.Printf("WhenToRun %s", service.WhenToRun)
				runSequenceMap[strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[1], " ", "")] = service.Id
				log.Printf("numberOfService %s", numberOfService)
			}
			if service.WhenToRun == "INIT" {
				runId = service.Id
			}
			numberOfService++
		}
	}
	log.Printf("runSequenceMap %s", runSequenceMap)
	for i := 0; i < numberOfService; i++ {
		log.Printf("Each run", idServiceMap[runId])
		returnMessage, returnErr := http.RequestCall(idServiceMap[runId].Url, "POST", idServiceMap[runId].Parameters)
		if returnErr != nil {
			log.Fatalf("returnErr %s", returnErr)
		}
		log.Printf("returnMessage %s", returnMessage)

		runId = runSequenceMap[runId]
	}
	defer wg.Done()
	return "", nil
}

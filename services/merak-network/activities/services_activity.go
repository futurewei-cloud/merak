package activities

import (
	"context"
	"fmt"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
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
	var runIds []string
	numberOfService := 0
	for _, service := range services {
		if service.WhereToRun == "network" {
			idServiceMap[service.Id] = service
			if service.WhenToRun != "INIT" {
				log.Printf("WhenToRun %s", service.WhenToRun)
				if strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[1], " ", "") == "AFTER" {
					runSequenceMap[strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[1], " ", "")] = service.Id
					log.Printf("numberOfService %s", numberOfService)
				}
				if strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[1], " ", "") == "BEFORE" {
					runSequenceMap[service.Id] = strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[1], " ", "")
					log.Printf("numberOfService %s", numberOfService)
				}
			}
			if service.WhenToRun == "INIT" {
				runIds = append(runIds, service.Id)
			}
			numberOfService++
		}
	}
	log.Printf("runSequenceMap %s", runSequenceMap)

	for _, eachRunId := range runIds {
		currentId := eachRunId
		for {
			//returnMessage, returnErr := http.RequestCall(idServiceMap[currentId].Url, "POST", idServiceMap[currentId].Parameters)
			//if returnErr != nil {
			//	log.Fatalf("returnErr %s", returnErr)
			//}
			//log.Printf("returnMessage %s", returnMessage)
			nextKey, ok := runSequenceMap[currentId]
			log.Printf("nextKey %s", nextKey)
			if ok {
				fmt.Println("value: ", nextKey)
				currentId = nextKey
			} else {
				fmt.Println("key not found")
				break
			}
		}
	}
	return "", nil
}

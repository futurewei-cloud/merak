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

package activities

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	"github.com/futurewei-cloud/merak/services/merak-network/http"
)

func DoServices(ctx context.Context, services []*pb.InternalServiceInfo, wg *sync.WaitGroup, projectId string) (string, error) {
	log.Println("DoServices")
	//defer wg.Done()
	log.Printf("DoServices %s", services)
	idServiceMap := make(map[string]*pb.InternalServiceInfo)
	runSequenceMap := make(map[string]string)
	var runIds []string
	numberOfService := 0
	for _, service := range services {
		if service.WhereToRun == "network" {
			idServiceMap[service.Name] = service
			if service.WhenToRun != "INIT" {
				log.Printf("WhenToRun %s", service.WhenToRun)
				if strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[0], " ", "") == "AFTER" {
					runSequenceMap[strings.Split(service.WhenToRun, ":")[1]] = service.Name
					log.Printf("numberOfService %s", numberOfService)
				}
				if strings.ReplaceAll(strings.Split(service.WhenToRun, ":")[0], " ", "") == "BEFORE" {
					runSequenceMap[service.Name] = strings.Split(service.WhenToRun, ":")[1]
					log.Printf("numberOfService %s", numberOfService)
				}
			}
			if service.WhenToRun == "INIT" {
				runIds = append(runIds, service.Name)
			}
			numberOfService++
		}
	}
	log.Printf("runSequenceMap %s", runSequenceMap)

	for _, eachRunId := range runIds {
		currentId := eachRunId
		for {
			if idServiceMap[currentId].Cmd == "curl" {
				log.Println("ssh service")
				var headers []string
				var payload string
				for _, parameter := range idServiceMap[currentId].Parameters {
					if strings.Split(parameter, " ")[0] == "-H" {
						headers = append(headers, strings.ReplaceAll(strings.Split(parameter, " ")[1], "'", ""))
					}
					if strings.Split(parameter, " ")[0] == "-d" {
						payload = strings.ReplaceAll(strings.Split(parameter, " ")[1], "'", "")
					}
				}
				returnMessage, returnErr := http.RequestCall(idServiceMap[currentId].Url, strings.Split(idServiceMap[currentId].Parameters[0], " ")[0], payload, headers)
				if returnErr != nil {
					log.Printf("returnErr %s", returnErr)
					return "", returnErr
				}
				log.Printf("returnMessage %s", returnMessage)

			}
			if idServiceMap[currentId].Cmd == "ssh" {
				var shellCommand string
				for _, command := range idServiceMap[currentId].Parameters {
					shellCommand = shellCommand + " " + command
				}
				cmd := exec.Command(shellCommand)
				stdout, err := cmd.Output()
				if err != nil {
					fmt.Println(err.Error())
					return "", err
				}
				log.Printf("shellCommand out: %s", string(stdout))
				log.Println("curl service")
			}

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

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
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Not enough arguments")
	}
	numCalls, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("numCalls: %d is not a valid number!\n", numCalls)
	}
	rps, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("RPS: %d is not a valid number!\n", rps)
	}

	ctx := context.Background()
	waitGroup := sync.WaitGroup{}
	limit := rate.NewLimiter(rate.Limit(rps), rps)
	begin := time.Now()
	for i := 0; i < numCalls; i++ {
		waitGroup.Add(1)
		go func() {
			id := uuid.New().String()[:10]
			defer waitGroup.Done()
			limit.Wait(ctx)
			log.Println("Adding tap " + id + " to br-int!")
			cmd := exec.Command("bash", "-c", "ovs-vsctl add-port br-int "+id+" --  set Interface "+id+" type=internal")
			stdout, err := cmd.Output()
			if err != nil {
				log.Println("ovs-vsctl failed for id " + id + string(stdout))
			}
		}()
	}
	waitGroup.Wait()
	end := time.Now()
	diff := end.Sub(begin)
	log.Println("Finished ", numCalls, " ovs calls. Time elapsed is ", diff.Milliseconds(), " ms")
}

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
package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	test_pb "github.com/futurewei-cloud/merak/api/proto/v1/ntest"
)

func caseTestCreatePing(ctx context.Context, in *pb.InternalTestConfig) (*pb.AgentReturnTestInfo, error) {

	log.Println("Create Test")

	// Create Device
	log.Println("Pinging " + in.DestIp)
	cmd := exec.Command("bash", "-c", "ping -c 1 "+in.DestIp)
	err := cmd.Run()
	var exerr *exec.ExitError
	if errors.As(err, &exerr) {
		fmt.Printf("Ping failed!: %d\n", exerr.ExitCode())
		return &pb.AgentReturnTestInfo{
			ReturnMessage: "Ping failed! " + strconv.Itoa(exerr.ExitCode()),
			ReturnCode:    common_pb.ReturnCode_FAILED,
			Status:        test_pb.TestStatus_FAILED,
		}, err
	}

	return &pb.AgentReturnTestInfo{
		ReturnMessage: "Ping Success! " + strconv.Itoa(exerr.ExitCode()),
		ReturnCode:    common_pb.ReturnCode_OK,
		Status:        test_pb.TestStatus_PASSED,
	}, err
}

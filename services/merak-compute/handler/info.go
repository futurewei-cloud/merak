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
	"log"
	"strconv"

	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
)

func caseInfo(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.ReturnComputeMessage, error) {
	log.Println("Operation Info")

	ids := RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_VM_SET)
	if ids.Err() != nil {
		log.Println("Unable to get VM IDs from redis", ids.Err())

		return &pb.ReturnComputeMessage{
			ReturnCode:    commonPB.ReturnCode_FAILED,
			ReturnMessage: "Unable to get node IDs from redis",
		}, ids.Err()
	}
	vms := []*pb.InternalVMInfo{}
	log.Println("Success in getting VM IDs!")
	count := 0
	total := 0
	for _, vmID := range ids.Val() {
		status, _ := strconv.Atoi(RedisClient.HGet(ctx, vmID, "status").Val())
		if commonPB.Status(status) == commonPB.Status_DONE {
			count += 1
		}
		total += 1
	}
	returnVM := pb.InternalVMInfo{
		Id: strconv.Itoa(count) + " out of " + strconv.Itoa(total) + " done!",
	}
	log.Println(strconv.Itoa(count) + " out of " + strconv.Itoa(total) + " done!")
	vms = append(vms, &returnVM)
	return &pb.ReturnComputeMessage{
		ReturnCode:    commonPB.ReturnCode_OK,
		ReturnMessage: "Success!",
		Vms:           vms,
	}, nil
}

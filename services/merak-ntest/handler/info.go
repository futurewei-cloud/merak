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
	pb "github.com/futurewei-cloud/merak/api/proto/v1/ntest"
	constants "github.com/futurewei-cloud/merak/services/common"
)

func caseInfo(ctx context.Context, in *pb.InternalTestConfiguration) (*pb.ReturnTestMessage, error) {
	log.Println("Operation Info")
	ids := RedisClient.SMembers(ctx, constants.NTEST_VM_SET)
	if ids.Err() != nil {
		log.Println("Unable to get VM IDs from redis", ids.Err())

		return &pb.ReturnTestMessage{
			ReturnCode:    commonPB.ReturnCode_FAILED,
			ReturnMessage: "Unable to get VM IDs from redis",
		}, ids.Err()
	}
	vms := []*pb.InternalVMTestResult{}
	log.Println("Success in getting VM IDs!")
	for _, vmID := range ids.Val() {
		srcPort, err := RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "srcPort").Int()
		if err != nil {
			log.Println("Failed to find port for vm " + vmID)
			srcPort = -1
		}

		src := pb.InternalVMHostInfo{
			Id:   RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "id").Val(),
			Ip:   RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "ip").Val(),
			Name: RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "name").Val(),
			Port: int32(srcPort),
		}

		dest := pb.InternalVMHostInfo{
			Id:   RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "destID").Val(),
			Ip:   RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "dest").Val(),
			Name: RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "name").Val(),
			Port: int32(srcPort),
		}
		vm := pb.InternalVMTestResult{
			Src:  &src,
			Dest: &dest,
		}
		status, err := strconv.Atoi(RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "status").Val())
		if err != nil {
			log.Println("Failed to convert status string to int!", err)
			return &pb.ReturnTestMessage{
				ReturnCode:    commonPB.ReturnCode_FAILED,
				ReturnMessage: "Failed to convert status string to int!",
				Results:       vms,
			}, err
		}
		testType, err := strconv.Atoi(RedisClient.HGet(ctx, constants.TEST_PREFIX+vmID, "testType").Val())
		if err != nil {
			return &pb.ReturnTestMessage{
				ReturnCode:    commonPB.ReturnCode_FAILED,
				ReturnMessage: "Failed to convert testType string to int!",
				Results:       vms,
			}, err
		}
		vm.Status = pb.TestStatus(status)
		vm.TestType = pb.TestType(testType)
		vms = append(vms, &vm)
	}

	return &pb.ReturnTestMessage{
		ReturnCode:    commonPB.ReturnCode_OK,
		ReturnMessage: "Success!",
		Results:       vms,
	}, nil
}

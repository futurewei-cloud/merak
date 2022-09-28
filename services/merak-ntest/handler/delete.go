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

	commonPB "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/ntest"
	constants "github.com/futurewei-cloud/merak/services/common"
)

func caseDelete(ctx context.Context, in *pb.InternalTestConfiguration) (*pb.ReturnTestMessage, error) {

	log.Println("Operation Delete")

	ids := RedisClient.SMembers(ctx, constants.NTEST_VM_SET)
	if ids.Err() != nil {
		log.Println("Unable to get VM IDs from redis", ids.Err())

		return &pb.ReturnTestMessage{
			ReturnCode:    commonPB.ReturnCode_FAILED,
			ReturnMessage: "Unable to get VM IDs from redis",
		}, ids.Err()
	}
	for _, vmID := range ids.Val() {
		RedisClient.HDel(ctx, vmID)
	}
	RedisClient.Del(ctx, constants.NTEST_VM_SET)
	return &pb.ReturnTestMessage{
		ReturnMessage: "Successfully started all create workflows!",
		ReturnCode:    commonPB.ReturnCode_OK,
	}, nil
}

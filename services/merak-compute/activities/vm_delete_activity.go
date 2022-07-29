package activities

import (
	"context"
	"strconv"
	"strings"

	agent_pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	compute_pb "github.com/futurewei-cloud/merak/api/proto/v1/compute"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func VmDelete(ctx context.Context) (*compute_pb.ReturnMessage, error) {
	logger := activity.GetLogger(ctx)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_NODE_IP_SET)
	if ids.Err() != nil {
		logger.Error("Unable get VM IDs from redis", ids.Err())

		return &compute_pb.ReturnMessage{
			ReturnCode:    common_pb.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, ids.Err()
	}
	logger.Info("Success in getting Pod IDs! " + ids.String())
	var agent_address strings.Builder
	for _, podID := range ids.Val() {
		vmIDsList := common.RedisClient.LRange(ctx, "l"+podID, 0, -1)
		if vmIDsList.Err() != nil {
			logger.Error("Unable get node vmIDsList from redis", vmIDsList.Err())
			return &compute_pb.ReturnMessage{
				ReturnCode:    common_pb.ReturnCode_FAILED,
				ReturnMessage: "Unable get node vmIDsList from redis",
			}, vmIDsList.Err()
		}
		agent_address.Reset()
		agent_address.WriteString(common.RedisClient.HGet(ctx, podID, "ip").Val())
		agent_address.WriteString(":")
		agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
		conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Info("Failed to dial gRPC server address: "+agent_address.String(), err)
		}
		client := agent_pb.NewMerakAgentServiceClient(conn)
		logger.Info("VM Ids " + vmIDsList.String() + "\n")
		for _, vmID := range vmIDsList.Val() {
			port := agent_pb.InternalPortConfig{
				OperationType: common_pb.OperationType_DELETE,
				Name:          common.RedisClient.HGet(ctx, vmID, "name").Val(),
				Projectid:     common.RedisClient.HGet(ctx, vmID, "projectID").Val(),
				Deviceid:      common.RedisClient.HGet(ctx, vmID, "deviceID").Val(),
				Remoteid:      common.RedisClient.HGet(ctx, vmID, "remoteID").Val(),
			}
			resp, err := client.PortHandler(ctx, &port)
			if err != nil {
				logger.Error("Unable Delete vm ID " + common.RedisClient.HGet(ctx, vmID, "hostIP").Val() + "Reason: " + resp.GetReturnMessage() + "\n")
				return &compute_pb.ReturnMessage{
					ReturnCode:    common_pb.ReturnCode_FAILED,
					ReturnMessage: "Unable Delete vm ID " + common.RedisClient.HGet(ctx, vmID, "hostIP").Val() + "Reason: " + resp.GetReturnMessage() + "\n",
				}, nil
			}
		}
	}
	common.RedisClient.FlushAll(ctx)
	return &compute_pb.ReturnMessage{
		ReturnCode:    common_pb.ReturnCode_OK,
		ReturnMessage: "Success!",
	}, nil
}

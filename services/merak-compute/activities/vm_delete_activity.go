package activities

import (
	"context"
	"strconv"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func VmDelete(ctx context.Context) (*pb.ReturnMessage, error) {
	logger := activity.GetLogger(ctx)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_NODE_IP_SET)
	if ids.Err() != nil {
		logger.Error("Unable get VM IDs from redis", ids.Err())

		return &pb.ReturnMessage{
			ReturnCode:    pb.ReturnCode_FAILED,
			ReturnMessage: "Unable get node IDs from redis",
		}, ids.Err()
	}
	logger.Info("Success in getting Pod IDs! " + ids.String())
	var agent_address strings.Builder
	for _, podID := range ids.Val() {
		vmIDsList := common.RedisClient.LRange(ctx, podID, 0, -1)
		if vmIDsList.Err() != nil {
			logger.Error("Unable get node vmIDsList from redis", vmIDsList.Err())
			return &pb.ReturnMessage{
				ReturnCode:    pb.ReturnCode_FAILED,
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
		client := pb.NewMerakAgentServiceClient(conn)
		logger.Info("VM Ids " + vmIDsList.String() + "\n")
		for _, vmID := range vmIDsList.Val() {
			port := pb.InternalPortConfig{
				OperationType: pb.OperationType_DELETE,
				Name:          common.RedisClient.HGet(ctx, vmID, "name").Val(),
				Projectid:     common.RedisClient.HGet(ctx, vmID, "projectID").Val(),
				Deviceid:      common.RedisClient.HGet(ctx, vmID, "deviceID").Val(),
				Remoteid:      common.RedisClient.HGet(ctx, vmID, "remoteID").Val(),
			}
			resp, err := client.PortHandler(ctx, &port)
			if err != nil {
				logger.Error("Unable create vm ID " + common.RedisClient.HGet(ctx, vmID, "hostIP").Val() + "Reason: " + resp.GetReturnMessage() + "\n")
			}
		}
	}
	common.RedisClient.FlushAll(ctx)
	return &pb.ReturnMessage{
		ReturnCode: pb.ReturnCode_OK,
	}, nil
}

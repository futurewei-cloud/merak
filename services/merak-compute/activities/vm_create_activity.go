package activities

import (
	"context"
	"strconv"
	"strings"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-compute/common"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func VmCreate(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger = log.With(logger)
	ids := common.RedisClient.SMembers(ctx, constants.COMPUTE_REDIS_NODE_IP_SET)
	if ids.Err() != nil {
		logger.Error("Unable get node IDs from redis", ids.Err())
		return "Failed!", ids.Err()
	}
	logger.Info("Success in getting Node IDs!", "NodeIDs", ids.Val())

	for _, podID := range ids.Val() {
		vmIDsList := common.RedisClient.LRange(ctx, podID, 0, -1)
		if vmIDsList.Err() != nil {
			logger.Error("Unable get node vmIDsList from redis", vmIDsList.Err())
			return "Failed!", vmIDsList.Err()
		}
		logger.Info("Pod IDs", ids.Val())
		logger.Info("VM Ids", vmIDsList.Val())
		for _, vmID := range vmIDsList.Val() {

			var agent_address strings.Builder
			agent_address.WriteString(common.RedisClient.HGet(ctx, vmID, "hostip").Val())
			agent_address.WriteString(":")
			agent_address.WriteString(strconv.Itoa(constants.AGENT_GRPC_SERVER_PORT))
			ctx := context.Background()
			conn, err := grpc.Dial(agent_address.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logger.Info("Failed to dial gRPC server address!", agent_address.String(), err)
			}
			client := pb.NewMerakAgentServiceClient(conn)

			logger.Info("Looking up ID ", vmID)
			vm := common.RedisClient.HGetAll(ctx, vmID)
			if vm.Err() != nil {
				logger.Error("Unable get node VM from redis for", "vmID", vmID, vm.Err())
				return "Failed!", vm.Err()
			}
			logger.Info("Sending to agent", vm.Val())
			port := pb.InternalPortConfig{
				OperationType: pb.OperationType_CREATE,
				Name:          common.RedisClient.HGet(ctx, vmID, "name").Val(),
				Vpcid:         common.RedisClient.HGet(ctx, vmID, "vpc").Val(),
				Tenantid:      common.RedisClient.HGet(ctx, vmID, "tenantID").Val(),
				Projectid:     common.RedisClient.HGet(ctx, vmID, "projectID").Val(),
				Subnetid:      common.RedisClient.HGet(ctx, vmID, "subnet").Val(),
				Gw:            common.RedisClient.HGet(ctx, vmID, "gw").Val(),
				Sg:            common.RedisClient.HGet(ctx, vmID, "sg").Val(),
			}
			resp, err := client.PortHandler(ctx, &port)
			if err != nil {
				logger.Error("Unable create vm", vmID, common.RedisClient.HGet(ctx, vmID, "hostip").Val())
				return "Failed!", err
			}
			logger.Info("Response from agent at address", resp)
			defer conn.Close()
		}
	}

	return common.TEMPORAL_SUCESS_CODE, nil
}

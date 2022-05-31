package service

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"go.temporal.io/sdk/client"
)

var (
	Port = flag.Int("port", constants.TOPOLOGY_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.UnimplementedMerakTopologyServiceServer
}

var TemporalClient client.Client

func (s *Server) TopologyHandler(ctx context.Context, in *pb.InternalToplogyInfo) (*pb.InternalTopologyInfo, error) {
	log.Printf("Received from TopologyHandler")

	return in, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalTopologyInfo) (*pb.InternalTopologyInfo, error) {
	log.Printf("Received on TestHandler")
	return in, nil
}

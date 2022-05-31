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
	Port = flag.Int("port", constants.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

var TemporalClient client.Client

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received on ComputeHandler %s", in)

	return in, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received on TestHandler %s", in)
	return in, nil
}

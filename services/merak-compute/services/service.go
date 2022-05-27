package services

import (
	"context"
	"flag"
	"log"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"github.com/futurewei-cloud/merak/services/common"
)

var (
	Port = flag.Int("port", common.COMPUTE_GRPC_SERVER_PORT, "The server port")
)

type Server struct {
	pb.UnimplementedMerakComputeServiceServer
}

func (s *Server) ComputeHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received")
	return in, nil
}

func (s *Server) TestHandler(ctx context.Context, in *pb.InternalComputeConfigInfo) (*pb.InternalComputeConfigInfo, error) {
	log.Printf("Received")
	return in, nil
}

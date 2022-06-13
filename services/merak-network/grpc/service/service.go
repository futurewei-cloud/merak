package service

//package main

import (
	"context"
	"flag"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-network/activities"
	"log"
)

var (
	Port          = flag.Int("port", constants.NETWORK_GRPC_SERVER_PORT, "The server port")
	returnMessage = pb.ReturnMessage{
		ReturnCode:    pb.ReturnCode_FAILED,
		ReturnMessage: "Unintialized",
	}
)

type Server struct {
	pb.UnimplementedMerakNetworkServiceServer
}

func (s *Server) InternalNetConfigConfigurationHandler(ctx context.Context, in *pb.InternalNetConfigInfo) (*pb.ReturnMessage, error) {
	log.Printf("Received on NetworkHandler %s", in)
	// Parse input

	switch op := in.OperationType; op {
	case pb.OperationType_INFO:
		log.Println("Info")
	case pb.OperationType_CREATE:
		for _, services := range in.Configuration.Services {
			log.Println(services)
		}
		for _, compute := range in.Configuration.Computes {
			log.Println(compute)
		}
		for _, network := range in.Configuration.Networks {
			log.Println(network)
			ctx := context.TODO()
			go activities.VnetCreate(ctx, network)
		}
		for _, storage := range in.Configuration.Storages {
			log.Println(storage)
		}
		for _, extraInfo := range in.Configuration.ExtraInfo {
			log.Println(extraInfo)
		}
		returnMessage.ReturnCode = pb.ReturnCode_OK
		returnMessage.ReturnMessage = "NetworkHandler: OperationType_CREATE"
		return &returnMessage, nil
	case pb.OperationType_UPDATE:
		log.Println("Update")
	case pb.OperationType_DELETE:
		log.Println("Delete")
	default:
		log.Println("Unknown Operation")
		returnMessage.ReturnCode = pb.ReturnCode_FAILED
		returnMessage.ReturnMessage = "NetworkHandler: Unknown Operation"
		return &returnMessage, nil
	}

	return &returnMessage, nil
}

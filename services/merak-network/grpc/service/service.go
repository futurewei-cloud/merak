package service

//package main

import (
	"context"
	"flag"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-network/activities"
	"log"
	"sync"
	"time"
)

var (
	Port = flag.Int("port", constants.NETWORK_GRPC_SERVER_PORT, "The server port")
	//returnMessage = pb.ReturnMessage{
	//	ReturnCode:    pb.ReturnCode_FAILED,
	//	ReturnMessage: "Unintialized",
	//}

	returnNetworkMessage = pb.ReturnNetworkMessage{
		ReturnCode:       pb.ReturnCode_FAILED,
		ReturnMessage:    "returnNetworkMessage Unintialized",
		Vpcs:             nil,
		SecurityGroupIds: nil,
	}
)

type Server struct {
	pb.UnimplementedMerakNetworkServiceServer
}

func (s *Server) NetConfigHandler(ctx context.Context, in *pb.InternalNetConfigInfo) (*pb.ReturnNetworkMessage, error) {
	log.Printf("Received on NetworkHandler %s", in)
	log.Printf("OP type %s", in.GetOperationType())

	wg := new(sync.WaitGroup)
	// Parse input

	//switch op := in.OperationType; op {
	switch op := in.GetOperationType(); op {
	case pb.OperationType_INFO:
		log.Println("Info")
	case pb.OperationType_CREATE:
		ctx := context.TODO()
		// services
		log.Println(in.Config.Services)
		wg.Add(1)
		//go activities.DoServices(ctx, in.Config.Services, wg)
		go activities.DoServices(ctx, in.Config.GetServices(), wg)

		//compute info done
		log.Println(in.Config.Computes)
		wg.Add(1)
		//go activities.RegisterNode(ctx, in.Config.Computes, wg)
		go activities.RegisterNode(ctx, in.Config.GetComputes(), wg)
		//network info done
		log.Println(in.Config.Network)
		wg.Add(1)
		networkReturn := make(chan *pb.ReturnNetworkMessage)
		//go activities.VnetCreate(ctx, in.Config.Network, wg, networkReturn)
		go activities.VnetCreate(ctx, in.Config.GetNetwork(), wg, networkReturn)

		//// storage info
		//for _, storage := range in.Config.Storage {
		//	log.Println(storage)
		//}
		//// extra info
		//for _, extraInfo := range in.Config.ExtraInfo {
		//	log.Println(extraInfo)
		//}
		log.Println("Before Wait")
		//wg.Wait()
		time.Sleep(15 * time.Second)
		log.Println("After Wait")

		returnNetworkMessage.ReturnCode = pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_CREATE"
		returnNetworkMessage := <-networkReturn
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	case pb.OperationType_UPDATE:
		log.Println("Update")
	case pb.OperationType_DELETE:
		log.Println("Delete")
	default:
		log.Println("Unknown Operation")
		returnNetworkMessage.ReturnCode = pb.ReturnCode_FAILED
		returnNetworkMessage.ReturnMessage = "NetworkHandler: Unknown Operation"
		return &returnNetworkMessage, nil
	}

	return &returnNetworkMessage, nil
}

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

package service

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

	netConfigId := in.Config.GetNetconfigId()

	//wg := new(sync.WaitGroup)
	var wg sync.WaitGroup
	// Parse input

	switch op := in.GetOperationType(); op {
	case pb.OperationType_INFO:
		ctx := context.TODO()
		log.Println("Info")
		networkInfoReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		go activities.VnetInfo(ctx, netConfigId, &wg, networkInfoReturn)
		//wg.Wait()
		time.Sleep(5 * time.Second)
		returnNetworkMessage.ReturnCode = pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_INFO"
		returnNetworkMessage := <-networkInfoReturn
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	case pb.OperationType_CREATE:
		ctx := context.TODO()
		// services
		log.Println(in.Config.Services)
		wg.Add(1)
		//go activities.DoServices(ctx, in.Config.Services, wg)
		projectId := in.Config.Network.Vpcs[0].ProjectId
		go activities.DoServices(ctx, in.Config.GetServices(), &wg, projectId)

		//compute info done
		log.Println(in.Config.Computes)
		wg.Add(1)
		//go activities.RegisterNode(ctx, in.Config.Computes, wg)
		go activities.RegisterNode(ctx, in.Config.GetComputes(), &wg, projectId)
		//network info done
		log.Println(in.Config.Network)
		wg.Add(1)
		networkReturn := make(chan *pb.ReturnNetworkMessage)
		//go activities.VnetCreate(ctx, in.Config.Network, wg, networkReturn)
		go activities.VnetCreate(ctx, netConfigId, in.Config.GetNetwork(), &wg, networkReturn, projectId)

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
		time.Sleep(5 * time.Second)
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
		ctx := context.TODO()
		networkDeleteReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		go activities.VnetDelete(ctx, netConfigId, &wg, networkDeleteReturn)
		//wg.Wait()
		time.Sleep(5 * time.Second)
		returnNetworkMessage.ReturnCode = pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_DELETE"
		returnNetworkMessage := <-networkDeleteReturn
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	default:
		log.Println("Unknown Operation")
		returnNetworkMessage.ReturnCode = pb.ReturnCode_FAILED
		returnNetworkMessage.ReturnMessage = "NetworkHandler: Unknown Operation"
		return &returnNetworkMessage, nil
	}

	return &returnNetworkMessage, nil
}

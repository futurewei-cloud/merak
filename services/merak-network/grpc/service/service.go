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

//package main

import (
	"context"
	"flag"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-network/activities"
	"log"
	"sync"
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

	netConfigId := in.Config.GetNetconfigId()
	//projectId := in.Config.Network.SecurityGroups[0].ProjectId

	//wg := new(sync.WaitGroup)
	var wg sync.WaitGroup
	// Parse input

	//switch op := in.OperationType; op {
	switch op := in.GetOperationType(); op {
	case pb.OperationType_INFO:
		ctx := context.TODO()
		log.Println("Info")
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_INFO haha"
		returnNetworkMessage.ReturnCode = pb.ReturnCode_OK
		networkInfoReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		go func() {
			defer wg.Done()
			var vnetInfoReturn, err = activities.VnetInfo(ctx, netConfigId)
			if err != nil {
				returnNetworkMessage.ReturnCode = pb.ReturnCode_FAILED
			}
			networkInfoReturn <- vnetInfoReturn
			log.Printf("networkInfoReturn: %s", networkInfoReturn)
		}()

		returnNetworkMessage := <-networkInfoReturn
		wg.Wait()
		log.Println("after wg")
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	case pb.OperationType_CREATE:
		ctx := context.TODO()
		// services
		log.Println(in.Config.Services)
		wg.Add(1)
		projectId := in.Config.Network.Vpcs[0].ProjectId
		//go activities.DoServices(ctx, in.Config.GetServices(), &wg, projectId)
		go func() {
			defer wg.Done()
			activities.DoServices(ctx, in.Config.GetServices(), &wg, projectId)
		}()

		//compute info done
		log.Println(in.Config.Computes)
		wg.Add(1)
		//go activities.RegisterNode(ctx, in.Config.GetComputes(), &wg, projectId)
		go func() {
			defer wg.Done()
			activities.RegisterNode(ctx, in.Config.GetComputes(), &wg, projectId)
		}()
		wg.Wait()
		//network info done
		log.Println(in.Config.Network)
		wg.Add(1)
		networkReturn := make(chan *pb.ReturnNetworkMessage)
		//go activities.VnetCreate(ctx, netConfigId, in.Config.GetNetwork(), &wg, networkReturn, projectId)
		go func() {
			defer wg.Done()
			var vnetCreateReturn, err = activities.VnetCreate(ctx, netConfigId, in.Config.GetNetwork(), &wg, networkReturn, projectId)
			if err != nil {
				returnNetworkMessage.ReturnCode = pb.ReturnCode_FAILED
			}
			networkReturn <- vnetCreateReturn
			log.Printf("networkInfoReturn: %s", vnetCreateReturn)
		}()

		//// storage info
		//for _, storage := range in.Config.Storage {
		//	log.Println(storage)
		//}
		//// extra info
		//for _, extraInfo := range in.Config.ExtraInfo {
		//	log.Println(extraInfo)
		//}
		log.Println("Before Wait")

		returnNetworkMessage.ReturnCode = pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_CREATE"
		returnNetworkMessage := <-networkReturn
		wg.Wait()
		log.Println("After Wait")
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	case pb.OperationType_UPDATE:
		log.Println("Update")
	case pb.OperationType_DELETE:
		log.Println("Delete")
		ctx := context.TODO()
		networkDeleteReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		//go activities.VnetDelete(ctx, netConfigId, &wg, networkInfoReturn)
		go func() {
			defer wg.Done()
			var vnetDeleteReturn, err = activities.VnetDelete(ctx, netConfigId, &wg, networkDeleteReturn)
			if err != nil {
				returnNetworkMessage.ReturnCode = pb.ReturnCode_FAILED
			}
			networkDeleteReturn <- vnetDeleteReturn
			log.Printf("networkInfoReturn: %s", vnetDeleteReturn)
		}()
		wg.Wait()
		//time.Sleep(5 * time.Second)
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

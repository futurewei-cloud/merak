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
	"log"
	"sync"

	common_pb "github.com/futurewei-cloud/merak/api/proto/v1/common"
	pb "github.com/futurewei-cloud/merak/api/proto/v1/network"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/merak-network/activities"
)

var (
	Port = flag.Int("port", constants.NETWORK_GRPC_SERVER_PORT, "The server port")

	returnNetworkMessage = pb.ReturnNetworkMessage{
		ReturnCode:       common_pb.ReturnCode_FAILED,
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

	var wg sync.WaitGroup
	var ifAnyFailure bool
	ifAnyFailure = false
	var currentError error

	switch op := in.GetOperationType(); op {
	case common_pb.OperationType_INFO:
		log.Println("Info")
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_INFO haha"
		returnNetworkMessage.ReturnCode = common_pb.ReturnCode_OK
		networkInfoReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		go func() {
			defer wg.Done()
			var vnetInfoReturn, err = activities.VnetInfo(netConfigId)
			if err != nil {
				returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
				returnNetworkMessage.ReturnMessage = err.Error()
				ifAnyFailure = true
				currentError = err
			}
			networkInfoReturn <- vnetInfoReturn
			log.Printf("networkInfoReturn: %s", networkInfoReturn)
		}()

		returnNetworkMessage := <-networkInfoReturn
		for len(networkInfoReturn) > 0 {
			<-networkInfoReturn
		}
		wg.Wait()
		log.Printf("returnNetworkMessage %s", returnNetworkMessage)
		if ifAnyFailure {
			return nil, currentError
		}
		return returnNetworkMessage, nil
	case common_pb.OperationType_CREATE:
		// services
		log.Println(in.Config.Services)
		wg.Add(1)
		projectId := in.Config.Network.Vpcs[0].ProjectId
		go func() {
			defer wg.Done()
			var _, err = activities.DoServices(in.Config.GetServices())
			if err != nil {
				returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
				returnNetworkMessage.ReturnMessage = err.Error()
				ifAnyFailure = true
				currentError = err
			}
		}()

		//compute info done
		log.Println(in.Config.Computes)
		wg.Add(1)
		go func() {
			defer wg.Done()
			var _, err = activities.RegisterNode(in.Config.GetComputes(), netConfigId)
			if err != nil {
				returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
				returnNetworkMessage.ReturnMessage = err.Error()
				ifAnyFailure = true
				currentError = err
			}
		}()
		wg.Wait()
		//network info done
		log.Println(in.Config.Network)
		wg.Add(1)
		networkReturn := make(chan *pb.ReturnNetworkMessage)
		go func() {
			defer wg.Done()
			var vnetCreateReturn, err = activities.VnetCreate(netConfigId, in.Config.GetNetwork(), projectId)
			if err != nil {
				returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
				returnNetworkMessage.ReturnMessage = err.Error()
				ifAnyFailure = true
				currentError = err
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

		returnNetworkMessage.ReturnCode = common_pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_CREATE"
		returnNetworkMessage := <-networkReturn
		for len(networkReturn) > 0 {
			<-networkReturn
		}
		wg.Wait()
		if ifAnyFailure {
			return nil, currentError
		}
		log.Printf("networkCreateReturn returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	case common_pb.OperationType_UPDATE:
		log.Println("Update")
	case common_pb.OperationType_DELETE:
		log.Println("Delete")
		networkDeleteReturn := make(chan *pb.ReturnNetworkMessage)
		wg.Add(1)
		go func() {
			defer wg.Done()
			var vnetDeleteReturn, err = activities.VnetDelete(netConfigId, networkDeleteReturn)
			if err != nil {
				returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
				returnNetworkMessage.ReturnMessage = err.Error()
				ifAnyFailure = true
				currentError = err
			}
			networkDeleteReturn <- vnetDeleteReturn
			log.Printf("networkDeleteReturn: %s", vnetDeleteReturn)
		}()
		returnNetworkMessage.ReturnCode = common_pb.ReturnCode_OK
		returnNetworkMessage.ReturnMessage = "NetworkHandler: OperationType_DELETE"
		returnNetworkMessage := <-networkDeleteReturn
		for len(networkDeleteReturn) > 0 {
			<-networkDeleteReturn
		}
		wg.Wait()
		log.Printf("============== After networkDeleteReturn")
		if ifAnyFailure {
			return nil, currentError
		}
		log.Printf("networkDeleteReturn returnNetworkMessage %s", returnNetworkMessage)
		return returnNetworkMessage, nil
	default:
		log.Println("Unknown Operation")
		returnNetworkMessage.ReturnCode = common_pb.ReturnCode_FAILED
		returnNetworkMessage.ReturnMessage = "NetworkHandler: Unknown Operation"
		return &returnNetworkMessage, nil
	}

	return &returnNetworkMessage, nil
}

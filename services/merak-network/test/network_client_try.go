package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/merak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:40054", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	testInternalSecurityGroupRulelnfo := pb.InternalSecurityGroupRulelnfo{
		OperationType:  pb.OperationType_CREATE,
		Id:             "1",
		Name:           "InternalSecurityGroupRulelnfo",
		Description:    "InternalSecurityGroupRulelnfo_description",
		Ethertype:      "5",
		Direction:      "6",
		Protocol:       "7",
		PortRange:      "8",
		RemoteGroupId:  "9",
		RemoteIpPrefix: "10",
	}
	testInternalRouterInfo := pb.InternalRouterInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "InternalRouterInfo",
		Subnets:       []string{"10.6.0.0/16", "10.7.0.0/16"},
	}
	testInternalGatewayInfo := pb.InternalGatewayInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "InternalGatewayInfo",
		Ips:           []string{"10.6.0.1", "10.7.0.1"},
	}
	testInternalSecurityGroupInfo := pb.InternalSecurityGroupInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "InternalSecurityGroupInfo",
		TenantId:      "123456789",
		ProjectId:     "123456789",
		Rules:         []*pb.InternalSecurityGroupRulelnfo{&testInternalSecurityGroupRulelnfo},
		ApplyTo:       []string{"ApplyTo"},
	}
	testInternalSubnetInfo := pb.InternalSubnetInfo{
		SubnetId:   "SubnetId1",
		SubnetCidr: "10.6.0.0/16",
		SubnetGw:   "10.6.0.1",
		NumberVms:  0,
	}
	testInternalSubnetInfo2 := pb.InternalSubnetInfo{
		SubnetId:   "SubnetId2",
		SubnetCidr: "10.7.0.0/16",
		SubnetGw:   "10.7.0.1",
		NumberVms:  0,
	}
	testInternalVpcInfo := pb.InternalVpcInfo{
		VpcId:     "VpcId1",
		TenantId:  "123456789",
		ProjectId: "123456789",
		Subnets:   []*pb.InternalSubnetInfo{&testInternalSubnetInfo, &testInternalSubnetInfo2},
	}
	testInternalNetworkInfo := pb.InternalNetworkInfo{
		OperationType:          pb.OperationType_CREATE,
		Id:                     "1",
		Name:                   "InternalNetworkInfo",
		NumberOfVpcs:           1,
		NumberOfSubnetPerVpc:   2,
		Vpcs:                   []*pb.InternalVpcInfo{&testInternalVpcInfo},
		NumberOfSecurityGroups: 1,
		Routers:                []*pb.InternalRouterInfo{&testInternalRouterInfo},
		Gateways:               []*pb.InternalGatewayInfo{&testInternalGatewayInfo},
		SecurityGroups:         []*pb.InternalSecurityGroupInfo{&testInternalSecurityGroupInfo},
	}
	testInternalServiceInfo := pb.InternalServiceInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "InternalServiceInfo",
		Cmd:           "InternalServiceInfo CMD",
		Url:           "InternalServiceInfo URL",
		Parameters:    []string{"Parameters"},
		ReturnCode:    nil,
		ReturnString:  []string{"ReturnString"},
		WhenToRun:     "WhenToRun",
		WhereToRun:    "WhereToRun",
	}
	testInternalComputeInfo := pb.InternalComputeInfo{
		OperationType: pb.OperationType_CREATE,
		Id:            "1",
		Name:          "InternalComputeInfo",
		Ip:            "IP",
		Mac:           "Mac",
		Veth:          "Veth",
	}
	testInternalStorageInfo := pb.InternalStorageInfo{Info: "InternalStorageInfo"}
	testInternalNetConfigExtraInfo := pb.InternalNetConfigExtraInfo{Info: "InternalNetConfigExtraInfo"}
	testInternalNetConfigConfiguration := pb.InternalNetConfigConfiguration{
		FormatVersion:  0,
		RevisionNumber: 0,
		RequestId:      "InternalNetConfigConfiguration RequestId",
		NetconfigId:    "InternalNetConfigConfiguration NetconfigId",
		MessageType:    0,
		Services:       []*pb.InternalServiceInfo{&testInternalServiceInfo},
		Computes:       []*pb.InternalComputeInfo{&testInternalComputeInfo},
		Network:        &testInternalNetworkInfo,
		Storage:        &testInternalStorageInfo,
		ExtraInfo:      &testInternalNetConfigExtraInfo,
	}
	testInternalNetConfigInfo := pb.InternalNetConfigInfo{
		OperationType: pb.OperationType_CREATE,
		Config:        &testInternalNetConfigConfiguration,
	}

	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// c := pb.NewGreeterClient(conn)
	c := pb.NewMerakNetworkServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	r, err := c.InternalNetConfigConfigurationHandler(ctx, &testInternalNetConfigInfo)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	//log.Printf("Greeting: %s", r.GetMessage())
	log.Printf("Greeting: %s", r.ReturnMessage)
}

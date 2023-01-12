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
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	pb "github.com/futurewei-cloud/merak/api/proto/v1/agent"
	constants "github.com/futurewei-cloud/merak/services/common"
	"github.com/futurewei-cloud/merak/services/common/logger"
	"github.com/futurewei-cloud/merak/services/common/metrics"

	"github.com/futurewei-cloud/merak/services/merak-agent/handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	gRPCPort       = flag.Int("gRPC port", constants.AGENT_GRPC_SERVER_PORT, "The server port")
	prometheusPort = flag.Int("Prometheus port", constants.PROMETHEUS_PORT, "The server port")
)

func main() {

	_, ok := os.LookupEnv(constants.MODE_ENV)
	if !ok {
		startPlugin()
	}
	var err error
	val, ok := os.LookupEnv(constants.MODE_ENV)
	if !ok {
		log.Println("No log level specified. Defaulting to INFO")
		val = "INFO"
	}
	handler.MerakLogger, err = logger.NewLogger(logger.LevelEnvParser(val), "merak-agent")
	if err != nil {
		handler.MerakLogger.Fatal("Failed to create logger\n", err)
	}
	// Start gRPC Server
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *gRPCPort))
	if err != nil {
		handler.MerakLogger.Fatal("ERROR: Failed to listen", err)
	}

	enforcement := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}
	kpServerParam := keepalive.ServerParameters{
		Time:    30 * time.Second,
		Timeout: 90 * time.Second,
	}

	gRPCServer := grpc.NewServer(
		grpc.MaxSendMsgSize(constants.GRPC_MAX_SEND_MSG_SIZE),
		grpc.MaxRecvMsgSize(constants.GRPC_MAX_RECV_MSG_SIZE),
		grpc.KeepaliveEnforcementPolicy(enforcement),
		grpc.KeepaliveParams(kpServerParam))

	go func() {
		hostname, err := os.Hostname()
		if err != nil {
			handler.MerakLogger.Fatal("Unable to get hostname!\n")
		}
		// Prometheus no hyphens allowed
		hostname = "vhost" + strings.Split(hostname, "-")[1]

		handler.PrometheusRegistry = prometheus.NewRegistry()
		handler.MerakMetrics = metrics.NewMetrics(handler.PrometheusRegistry, "merak_agent_"+hostname)
		http.Handle("/metrics", promhttp.HandlerFor(
			handler.PrometheusRegistry,
			promhttp.HandlerOpts{Registry: handler.PrometheusRegistry}))
		http.ListenAndServe(fmt.Sprintf(":%d", *prometheusPort), nil)
	}()

	pb.RegisterMerakAgentServiceServer(gRPCServer, &handler.Server{})
	handler.MerakLogger.Info("Starting gRPC server. Listening at %v", lis.Addr())
	if err := gRPCServer.Serve(lis); err != nil {
		handler.MerakLogger.Fatal("failed to serve: %v\n", err)
	}
}

func startPlugin() {
	if len(os.Args) < 3 {
		handler.MerakLogger.Fatal("Not enough arguments\n")
	}
	remote_server := os.Args[1]
	if net.ParseIP(remote_server) == nil {
		handler.MerakLogger.Fatal("Invalid IP address %s\n", remote_server)
	}
	handler.RemoteServer = remote_server
	remote_port := os.Args[2]
	remote_port_int, err := strconv.Atoi(os.Args[2])

	if err != nil {
		handler.MerakLogger.Fatal("Port: %d is not a valid number!\n", remote_port_int)
	}
	if remote_port_int > constants.MAX_PORT || remote_port_int < constants.MIN_PORT {
		handler.MerakLogger.Fatal("Port: %d is not within a valid range!\n", remote_port_int)
	}
	cmdString := "service rsyslog restart && /etc/init.d/openvswitch-switch restart && /merak-bin/AlcorControlAgent -d -a " + remote_server + " -p " + remote_port
	handler.MerakLogger.Info("Executing command " + cmdString)
	// Start plugin
	cmd := exec.Command("bash", "-c", cmdString)
	cmd.Dir = "/"
	err = cmd.Start()
	if err != nil {
		handler.MerakLogger.Fatal("failed to start plugin", err)
	}
	handler.MerakLogger.Info("Started ACA %d\n", cmd.Process.Pid)
}

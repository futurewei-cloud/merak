package main

import (
	"flag"
	"fmt"
	"net"
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf())

}

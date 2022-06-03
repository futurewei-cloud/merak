module := merak-network

merak.network:: main

main:
	go build -o services/merak-network/build/merak-network services/merak-netowrk/network.go
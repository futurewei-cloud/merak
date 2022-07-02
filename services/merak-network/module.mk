module := merak-network

merak.network:: network

network:
	go build -o services/merak-network/merak-network services/merak-network/network.go

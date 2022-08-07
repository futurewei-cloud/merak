module := merak-network

merak.network:: network

network:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/merak-network/merak-network services/merak-network/network.go

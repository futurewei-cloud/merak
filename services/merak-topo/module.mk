module := merak-topo

merak.topo:: topo 

topo:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/merak-topo/build/merak-topo services/merak-topo/main.go
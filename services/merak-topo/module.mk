module := merak-topo

merak.topo:: main 

main:
	go build -o services/merak-topo/build/merak-topo services/merak-topo/main.go
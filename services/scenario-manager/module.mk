module := scenario-manager

scenario.manager:: scenario 

scenario:
	go build -o services/scenario-manager/build/scenario-manager services/scenario-manager/main.go
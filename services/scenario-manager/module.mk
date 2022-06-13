module := scenario-manager

scenario.manager:: main 

main:
	go build -o services/scenario-manager/build/scenario-manager services/scenario-manager/main.go
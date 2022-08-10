module := scenario-manager

scenario.manager:: scenario 

scenario:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/scenario-manager/build/scenario-manager services/scenario-manager/main.go
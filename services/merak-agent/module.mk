module := merak-agent

merak-agent: agent

agent:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o services/merak-agent/build/merak-agent services/merak-agent/agent.go

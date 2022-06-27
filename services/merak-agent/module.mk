module := merak-agent

merak-agent: agent

agent:
	go build -o services/merak-agent/build/merak-agent services/merak-agent/agent.go

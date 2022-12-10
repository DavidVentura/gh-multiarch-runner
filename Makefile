PHONY: build

dist/agent-amd64:
	GOARCH=amd64 go build -o dist/agent-amd64 cmd/agent/main.go
dist/agent-arm64:
	GOARCH=arm64 go build -o dist/agent-arm64 cmd/agent/main.go
dist/agent-riscv64:
	GOARCH=riscv64 go build -o dist/agent-riscv64 cmd/agent/main.go

dist/orchestrator: 
	go build -o dist/orchestrator cmd/orchestrator/main.go
build: dist/agent-amd64 dist/agent-arm64 dist/agent-riscv64 dist/orchestrator
	:

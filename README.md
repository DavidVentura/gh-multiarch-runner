
POC repo for self-hosted, on-demand ephemeral GH runners on multiple architectures

Workers


```
GH -> Orchestrator [Fetch creation token] -> Node orchestrator
```

```
Node Orchestrator -> Spins up firecracker with Actions image with --ephemeral
```

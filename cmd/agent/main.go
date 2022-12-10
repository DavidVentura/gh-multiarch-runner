package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"github.com/davidventura/gh-multiarch-runner/pkgs/agent"
)

func main() {
	var wg sync.WaitGroup
	ag := new(agent.Agent)
	rpc.Register(ag)
	port := 2345
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	wg.Add(1)
	go func() {
		fmt.Printf("Listening on port %d for RPC requests\n", port)
		for {
			rpc.Accept(l)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		agent.ProcessWorkQueue()
		wg.Done()
	}()
	wg.Wait()
}

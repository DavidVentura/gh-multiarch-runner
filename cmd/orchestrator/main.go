package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"sync"

	"github.com/davidventura/gh-multiarch-runner/pkgs/agent"
	"github.com/davidventura/gh-multiarch-runner/pkgs/gh"
	"golang.org/x/exp/slices"
)

var t gh.AppToken

type Worker struct {
	Labels []string
	Client *rpc.Client
}

type WorkRequest struct {
	wh          gh.WebHookEvent
	runnerToken gh.RunnerToken
}

var workRequests = make(chan WorkRequest, 1)
var workers = []Worker{}

func (w Worker) CanProcessLabels(labels []string) bool {
	for _, wantedLabel := range labels {
		if !slices.Contains(w.Labels, wantedLabel) {
			return false
		}
	}
	return true
}

func (w Worker) SendJob(repoName string, runnerToken string) {

	fmt.Println("Should now send job")
	var ret = struct{}{}
	err := w.Client.Call("Agent.Work", agent.AgentRequest{RepoName: repoName, RunnerToken: runnerToken}, &ret)
	if err != nil {
		log.Fatal("Work call:", err)
	}
}

func MakeWorker(serverAddress string, port int) Worker {

	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", serverAddress, port))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	labels := struct {
		Labels []string
	}{}
	err = client.Call("Agent.Labels", struct{}{}, &labels)
	if err != nil {
		log.Fatal("Labels call:", err)
	}
	fmt.Printf("Creating worker with labels %v\n", labels)
	return Worker{Client: client, Labels: labels.Labels}

}
func handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("ERROR %v\n", err)
		return
	}
	wh := gh.WebHookEvent{}
	json.Unmarshal(body, &wh)
	runnerToken, err := gh.GetInstallationToken(wh.Installation.ID, t.Token)
	if err != nil {
		fmt.Printf("ERROR %v\n", err)
		return
	}
	if wh.Action == "queued" {
		workRequests <- WorkRequest{wh: wh, runnerToken: runnerToken}
	}
}

func processRequests() {
	fmt.Println("Processing requests..")
	for workRequest := range workRequests {
		fmt.Printf("For WH of repo %s and installation %d with labels %v\n",
			workRequest.wh.Repository.FullName, workRequest.wh.Installation.ID,
			workRequest.wh.WorkflowJob.Labels)

		sent := false
		for _, w := range workers {
			if !w.CanProcessLabels(workRequest.wh.WorkflowJob.Labels) {
				continue
			}
			w.SendJob(workRequest.wh.Repository.FullName, workRequest.runnerToken.Token)
			sent = true
			break
		}
		if !sent {
			fmt.Println("Did not find any suitable worker. Dropping.")
		}
	}
}

func main() {
	t = gh.MakeAppToken()
	http.HandleFunc("/", handler)
	workers = append(workers, MakeWorker("192.168.2.101", 2345)) // x64
	/*
		workers = append(workers, MakeWorker("192.168.2.171", 2345)) // aarch64
		workers = append(workers, MakeWorker("192.168.2.118", 2345)) // riscv64
	*/
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fmt.Println("Waiting for webhooks...")
		http.ListenAndServe(":9999", nil)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		processRequests()
		wg.Done()
	}()
	wg.Wait()

}

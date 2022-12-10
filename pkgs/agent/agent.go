package agent

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

var workRequests = make(chan AgentRequest, 1)
var labels = []string{strings.ReplaceAll(runtime.GOARCH, "amd64", "x64")}

type Agent struct {
	labels []string
}

type Labels struct {
	Labels []string
}

type AgentRequest struct {
	RepoName    string
	RunnerToken string
}

func (a *Agent) Work(req *AgentRequest, _ *struct{}) error {
	fmt.Printf("Working %v\n", req)
	workRequests <- *req
	return nil
}

func (a *Agent) Labels(_ *struct{}, labels *Labels) error {
	l := []string{strings.ReplaceAll(runtime.GOARCH, "amd64", "x64")}
	fmt.Printf("Returning %s\n", l)
	*labels = Labels{Labels: l}
	return nil
}
func ProcessWorkQueue() {
	fmt.Println("Processing work queue...")
	for wr := range workRequests {
		fmt.Printf("Should now work on repo %s with token %s\n", wr.RepoName, wr.RunnerToken)
		cmd := exec.Command("./github-act-runner", "configure",
			"--url", fmt.Sprintf("https://github.com/%s", wr.RepoName),
			"--name", "somename3",
			"--token", wr.RunnerToken,
			"--labels", strings.Join(labels, ","),
			"--unattended",
			"--ephemeral")
		cmd.Dir = "/home/david/actions-runner"
		fmt.Println("Command 1 going..")
		combined, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Failed to create command.. %s", err)
			log.Printf("output.. \n%s\n", string(combined))
			continue
		}

		cmd = exec.Command("./github-act-runner", "run")
		cmd.Dir = "/home/david/actions-runner"
		fmt.Println("Going to run..")
		combined, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Failed to create command.. %s", err)
			log.Printf("output.. \n%s\n", string(combined))
			continue
		}
	}
}

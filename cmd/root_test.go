package cmd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/k-kinzal/progressived/cmd"
	"github.com/k-kinzal/progressived/cmd/cli"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

var (
	client  = &cli.Client{Scheme: "http", Host: "localhost", Port: 9000}
	serveCh = make(chan error, 1)
)

func Serve(port int, debug bool) {
	os.Args = []string{"gotest"}
	if port > 0 {
		os.Args = append(os.Args, "--port", strconv.Itoa(port))
	}
	if debug {
		os.Args = append(os.Args, "--debug")
	}

	err := cmd.Execute()
	serveCh <- err
}

func WaitServe() {
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port))
		if err != nil {
			continue
		}
		conn.Close()
		return
	}
}

func Shutdown(t *testing.T, signum syscall.Signal) {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	syscall.Kill(syscall.Getpid(), signum)
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				t.Fatal("shutdown did not complete within 1 second")
			}
			return
		case err := <-serveCh:
			if err != nil {
				t.Fatal(err)
			}
			return
		}
	}
}

func TestExecute(t *testing.T) {
	go Serve(client.Port, false)
	WaitServe()
	Shutdown(t, syscall.SIGUSR1) // force shutdown
}

func TestExecute_gracefulShutdown(t *testing.T) {
	go Serve(client.Port, false)
	WaitServe()
	Shutdown(t, syscall.SIGUSR2) // graceful shutdown
}

func TestExecute_putDeployment(t *testing.T) {
	go Serve(client.Port, false)
	defer Shutdown(t, syscall.SIGUSR1)

	WaitServe()

	var req *request.PutDeploymentRequest
	body := `
{
	"interval": 60,
	"provider": {
		"type": "inmemory"
	},
	"step": {
		"algorithm": "increase",
		"threshold": 25
	}
}
`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}

	if _, _, err := client.PutDeployment("123", req); err != nil {
		t.Fatal(err)
	}
}

func TestExecute_getDeployment(t *testing.T) {
	go Serve(client.Port, false)
	defer Shutdown(t, syscall.SIGUSR1)

	WaitServe()

	if _, _, err := client.DescribeDeployment("123", &request.DescribeDeploymentRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestExecute_listDeployment(t *testing.T) {
	go Serve(client.Port, false)
	defer Shutdown(t, syscall.SIGUSR1)

	WaitServe()

	if _, _, err := client.ListDeployments(&request.ListDeploymentRequest{}); err != nil {
		t.Fatal(err)
	}
}

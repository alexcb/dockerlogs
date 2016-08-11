package main

import (
	"bufio"
	"fmt"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

// curl --unix-socket /var/run/docker.sock 'http:/containers/1a210a4481b7/logs?stderr=1&stdout=1&timestamps=1&follow=1'

type containerLogs struct {
	ID string
	ch chan string
}

func main() {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	containerLogsList := []containerLogs{}

	for _, c := range containers {
		ch := make(chan string, 1000)
		containerLogsList = append(containerLogsList, containerLogs{
			c.ID,
			ch,
		})
		fmt.Println(c.ID)
		go tailLogs(c.ID, ch)
	}

	for _, c := range containerLogsList {
		x := <-c.ch
		fmt.Printf("got: %v\n", x)
	}

}

func tailLogs(containerID string, ch chan<- string) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	body, err := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})

	reader := bufio.NewReader(body)

	x, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	ch <- x
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

// curl --unix-socket /var/run/docker.sock 'http:/containers/1a210a4481b7/logs?stderr=1&stdout=1&timestamps=1&follow=1'

type logLine struct {
	timestamp string
	line      string
}

type containerLogs struct {
	ID   string
	line *logLine
	ch   chan logLine
}

type logtail struct {
	containerLogsList []containerLogs
}

func NewLogTail() *logtail {
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
		ch := make(chan logLine, 1000)
		containerLogsList = append(containerLogsList, containerLogs{
			ID:   c.ID,
			line: nil,
			ch:   ch,
		})
		fmt.Println(c.ID)
		go tailLogs(c.ID, ch)
	}

	return &logtail{
		containerLogsList,
	}
}

func (s *logtail) getLine() *logLine {
	for {
		for i, _ := range s.containerLogsList {
			c := &s.containerLogsList[i]
			if c.line == nil {
				select {
				case x := <-c.ch:
					c.line = &x
				default:
				}
			}
		}

		mini := -1
		for i, _ := range s.containerLogsList {
			c := &s.containerLogsList[i]
			if c.line != nil {
				if mini == -1 {
					mini = i
				} else {
					if strings.Compare(s.containerLogsList[mini].line.timestamp, c.line.timestamp) < 0 {
						mini = i
					}
				}
			}
		}

		var line *logLine
		if mini >= 0 {
			c := &s.containerLogsList[mini]
			line = c.line
			c.line = nil
		}

		if line != nil {
			return line
		}
		//time.Sleep(time.Millisecond)
	}
}

func main() {

	lt := NewLogTail()

	time.Sleep(5 * time.Millisecond)

	for {
		line := lt.getLine()
		fmt.Printf("got: %v\n", line)
	}

}

func tailLogs(containerID string, ch chan<- logLine) {
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

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Printf("exiting\n")
			return
		} else if err != nil {
			log.Fatalf("Failed to read container %v log: %v", containerID, err)
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		//TODO figure out whats in the first 8 bytes
		line = line[8:]

		x := strings.SplitN(line, " ", 2)
		ch <- logLine{
			timestamp: x[0],
			line:      x[1],
		}
	}

}

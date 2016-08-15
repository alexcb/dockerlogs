package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"

	"github.com/aybabtme/humanlog"
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

func (s *logtail) readFromChannels() {
	// grab latest log from each channel (if available)
	numEmptyChannels := 0
	for i, _ := range s.containerLogsList {
		c := &s.containerLogsList[i]
		if c.line == nil {
			select {
			case x := <-c.ch:
				c.line = &x
			default:
				numEmptyChannels++
			}
		}
	}

	// if no logs are available, block until at least one is
	if numEmptyChannels == len(s.containerLogsList) {
		cases := make([]reflect.SelectCase, len(s.containerLogsList))
		for i, _ := range s.containerLogsList {
			ch := s.containerLogsList[i].ch
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
		}

		// Block
		chosen, value, ok := reflect.Select(cases)

		if !ok {
			panic("not ok channel from select")
		}

		logLine := value.Interface().(logLine)
		s.containerLogsList[chosen].line = &logLine
	}
}

func (s *logtail) getLine() *logLine {
	for {
		s.readFromChannels()

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

		if mini >= 0 {
			c := &s.containerLogsList[mini]
			line := c.line
			c.line = nil
			return line
		}
	}
}

func main() {

	lt := NewLogTail()

	// Sleep to make sure all files have been read by the corresponding thread
	time.Sleep(10 * time.Millisecond)

	opts := humanlog.DefaultOptions

	//logrusEntry := LogrusHandler{Opts: opts}
	jsonEntry := humanlog.JSONHandler{Opts: opts}

	dst := os.Stdout

	for {
		line := lt.getLine()
		lineData := []byte(line.line)

		switch {

		case jsonEntry.TryHandle(lineData):
			dst.Write(jsonEntry.Prettify(false))

		//case logrusEntry.CanHandle(line) && logfmt.Parse(lineData, true, true, logrusEntry.visit):
		//	dst.Write(logrusEntry.Prettify(false))

		default:
			dst.Write(lineData)
		}
		dst.Write([]byte("\n"))

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

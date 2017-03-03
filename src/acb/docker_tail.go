package dockerlogs

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

type logLine struct {
	Timestamp time.Time
	Line      string
}

type containerLogs struct {
	ID   string
	Name string
	line *logLine
	ch   chan logLine
}

type logtail struct {
	containerLogsList []containerLogs
}

func NewLogTail(cli *client.Client) *logtail {

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
			Name: strings.TrimPrefix(c.Names[0], "/"),
			line: nil,
			ch:   ch,
		})
		fmt.Println(c.ID)
		go tailDockerLog(c.ID, ch)
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

func (s *logtail) GetLine() (string, *logLine) {
	for {
		s.readFromChannels()

		mini := -1
		for i, _ := range s.containerLogsList {
			c := &s.containerLogsList[i]
			if c.line != nil {
				if mini == -1 {
					mini = i
				} else {
					if c.line.Timestamp.Before(s.containerLogsList[mini].line.Timestamp) {
						mini = i
					}
				}
			}
		}

		if mini >= 0 {
			c := &s.containerLogsList[mini]
			line := c.line
			c.line = nil
			return c.Name, line
		}
	}
}

func tailDockerLog(containerID string, ch chan<- logLine) {
	cli := MustGetDockerCli()

	body, _ := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
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
		//double TODO, sometimes a line is truncated, then the next line contains less than 8 bytes before the timestamp
		//so we will just remove all bytes until we find a '2' as in year '2004', this is terrible.
		for {
			if len(line) < 5 {
				break
			}

			if line[0] == '2' && line[4] == '-' {
				break
			}
			line = line[1:]
		}
		if len(line) < 10 {
			continue
		}

		x := strings.SplitN(line, " ", 2)

		timestamp, err := time.Parse(time.RFC3339Nano, x[0])
		if err != nil {
			log.Fatalf("Failed to parse timestamp %s: %s\n", x[0], err)
		}

		ch <- logLine{
			Timestamp: timestamp,
			Line:      strings.Trim(x[1], " \n\t\r"),
		}
	}
}

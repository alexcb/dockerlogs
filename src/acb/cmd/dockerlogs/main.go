package main

import (
	"fmt"
	"time"

	"acb"
)

// curl --unix-socket /var/run/docker.sock 'http:/containers/1a210a4481b7/logs?stderr=1&stdout=1&timestamps=1&follow=1'

func main() {

	cli := dockerlogs.MustGetDockerCli()
	lt := dockerlogs.NewLogTail(cli)

	maxContainerNameLength := dockerlogs.GetMaxContainerNameLength(cli)
	fmt.Printf("=--==- %v\n", maxContainerNameLength)

	// Sleep to make sure all files have been read by the corresponding thread
	time.Sleep(10 * time.Millisecond)

	for {
		containerName, line := lt.GetLine()
		timestamp := line.Timestamp.Format("2006-01-02T15:04:05")

		parsedLog := dockerlogs.ParseLog(line.Line)

		fmt.Printf("%s %s %s\n",
			dockerlogs.PadLeft(containerName, maxContainerNameLength),
			timestamp,
			parsedLog.Format())
	}

}

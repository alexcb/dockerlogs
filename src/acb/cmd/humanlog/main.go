package main

import (
	"acb"
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// curl --unix-socket /var/run/docker.sock 'http:/containers/1a210a4481b7/logs?stderr=1&stdout=1&timestamps=1&follow=1'

func main() {

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		text = strings.TrimSuffix(text, "\n")
		parsedLog := dockerlogs.ParseLog(text)

		fmt.Printf("%s\n", parsedLog.Format())
	}

}

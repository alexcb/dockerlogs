package dockerlogs

import (
	"strings"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

func MustGetDockerCli() *client.Client {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}
	return cli
}

func GetMaxContainerNameLength(cli *client.Client) int {
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	l := 0
	for _, c := range containers {
		n := strings.TrimPrefix(c.Names[0], "/")
		if len(n) > l {
			l = len(n)
		}
	}
	return l
}

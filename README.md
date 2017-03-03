# dockerlogs
docker log utility


export GOPATH=`pwd`
go run src/acb/cmd/dockerlogs/main.go
go build -o ~/bin/docker-logs src/acb/cmd/dockerlogs/main.go
go build -o ~/bin/humanlog src/acb/cmd/humanlog/main.go

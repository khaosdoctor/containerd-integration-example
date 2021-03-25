# Integration example for ContainerD and Golang

> Example application for integrating ContainerD with Golang following [this article](https://blog.lsantos.dev/integrando-containers-na-sua-aplicacao-com-containerd) from my blog

## Instructions

1. This needs to be executed within a machine with both [containerd](https://containerd.io/docs/getting-started/) and [runc](https://github.com/opencontainers/runc) installed
2. Only works for Linux machines

After cloning the repository in any directory of your linux machine:

1. `go get` to fetch the modules
2. `go build src/main.go` to build the binary
3. `sudo ./main` to run the application

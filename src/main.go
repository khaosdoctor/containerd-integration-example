package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func main() {
	if err := createAPI(); err != nil {
		log.Fatal(err)
	}
}

func createAPI () error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	if err != nil {
		return err
	}

	ctx := namespaces.WithNamespace(context.Background(), "lsantos")

	image, err := client.Pull(ctx, "docker.io/khaosdoctor/simple-node-api:latest", containerd.WithPullUnpack)
	if err != nil {
		return err
	}
	log.Printf("Imagem %q baixada", image.Name())

	container, err := createContainer(ctx, client, image)
	if err != nil {
		return err
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	task, err := createIOTask(ctx, container)
	if err != nil {
		return err
	}
	defer task.Delete(ctx)

	exitStatus, err := task.Wait(ctx)
	if err != nil {
		log.Println(err)
	}

	if err := task.Start(ctx); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return err
	}

	status := <-exitStatus
	exitCode, _, err := status.Result()
	if err != nil {
		return err
	}
	log.Printf("%q foi finalizado com status: %d\n", container.ID(), exitCode)

	return nil
}

func createContainer (
	ctx context.Context,
	client *containerd.Client,
	image containerd.Image,
) (containerd.Container, error) {

	hasher := sha256.New()
	hasher.Write([]byte(time.Now().String()))
	salt := hex.EncodeToString(hasher.Sum(nil))[0:8]

	containerName := "simple-api-" + salt
	log.Printf("Criando um novo container chamado %q", containerName)

	imageSpecs := containerd.WithNewSpec(
		oci.WithDefaultSpec(),
		oci.WithImageConfig(image),
		oci.WithEnv([]string{"PORT=8080"}),
		oci.WithHostNamespace(specs.NetworkNamespace),
		oci.WithHostHostsFile,
		oci.WithHostResolvconf,
		)

	container, err := client.NewContainer(
		ctx,
		containerName,
		containerd.WithNewSnapshot(containerName + "-snapshot", image),
		imageSpecs,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Criado novo container %q", containerName)
	return container, nil
}

func createIOTask (ctx context.Context, container containerd.Container) (containerd.Task, error) {
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, err
	}
	return task, nil
}

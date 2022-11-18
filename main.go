package main

import (
	"context"
	"fmt"
	"path"

	"dagger.io/dagger"
)

type Kapp struct {
}

func (Kapp) Deploy(
	ctx context.Context,
	app string,
	directory string,
	namespace string,
	url string,
	kubeconfig string,
) (string, error) {

	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	if namespace == "" {
		namespace = "default"
	}

	client, err := dagger.Connect(ctx)
	if err != nil {
		return "", err
	}

	workdir := client.Host().Workdir()

	file := workdir.Directory(directory)

	kubeconfigPath, kubeconfigFilename := path.Split(kubeconfig)
	kubeconfigSecret := client.Host().Directory(kubeconfigPath).File(kubeconfigFilename).Secret()

	container := client.Container().
		From("ghcr.io/vmware-tanzu/carvel-docker-image").
		WithMountedSecret("/root/.kube/config", kubeconfigSecret)

	execOpts := dagger.ContainerExecOpts{
		Args: []string{"kapp", "deploy", "-y", "-n", namespace, "-a", app, "-f"},
	}

	if url == "" {
		execOpts.Args = append(execOpts.Args, "/source")
		container = container.WithMountedDirectory("/source", file)
	} else {
		execOpts.Args = append(execOpts.Args, url)
	}

	output, err := container.Exec(execOpts).Stdout().Contents(ctx)
	if err != nil {
		return output, err
	}

	return output, nil

}

func main() {
	dagger.Serve(Kapp{})
}

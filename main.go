package main

import (
	"context"
	"fmt"

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
	fileName string,
	kubeconfig string,
) (string, error) {

	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	if namespace == "" {
		namespace = "default"
	}

	execOpts := dagger.ContainerExecOpts{
		Args: []string{"kapp", "deploy", "-y", "-n", namespace, "-a", app, "-f"},
	}

	client, err := dagger.Connect(ctx)
	if err != nil {
		return "", err
	}

	workdir := client.Host().Workdir()

	file := workdir.Directory(directory).File(fileName)

	kubeconfigSecret := workdir.Directory(".").File(kubeconfig).Secret()

	container := client.Container().
		From("ghcr.io/vmware-tanzu/carvel-docker-image").
		WithMountedSecret("/root/.kube/config", kubeconfigSecret)

	var output string
	if url == "" {
		execOpts.Args = append(execOpts.Args, "/source")
		output, err = container.
			WithMountedFile("/source", file).
			Exec(execOpts).
			Stdout().
			Contents(ctx)
	} else {
		execOpts.Args = append(execOpts.Args, url)
		output, err = container.Exec(execOpts).Stdout().Contents(ctx)
	}

	if err != nil {
		return "", err
	}

	return output, nil

}

func main() {
	dagger.Serve(Kapp{})
}

package main

import (
	"context"
	"fmt"

	"dagger.io/dagger/sdk/go/dagger"
)

type Kapp struct {
}

func (Kapp) LoadCredentials(ctx context.Context, kubeconfig string) (dagger.SecretID, error) {

	client, err := dagger.Connect(ctx)
	if err != nil {
		return "", err
	}

	secretId, err := client.Host().Workdir().Read().Directory(".").File(kubeconfig).Secret().ID(ctx)
	if err != nil {
		return "", err
	}

	return secretId, nil
}

func (Kapp) Deploy(ctx context.Context, app string, directory string, namespace string, url string, file string, kubeconfig dagger.SecretID) (string, error) {

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

	fileId, err := client.Host().Workdir().Read().Directory(directory).File(file).ID(ctx)
	if err != nil {
		return "", err
	}

	containerId, err := client.Container().From("ghcr.io/vmware-tanzu/carvel-docker-image").WithMountedSecret("/root/.kube/config", kubeconfig).ID(ctx)
	if err != nil {
		return "", err
	}

	containerOpts := dagger.ContainerOpts{
		ID: containerId,
	}

	var output string
	if url == "" {
		execOpts.Args = append(execOpts.Args, "/source")
		output, err = client.Container(containerOpts).WithMountedFile("/source", fileId).Exec(execOpts).Stdout().Contents(ctx)
	} else {
		execOpts.Args = append(execOpts.Args, url)
		output, err = client.Container(containerOpts).Exec(execOpts).Stdout().Contents(ctx)
	}

	if err != nil {
		return "", err
	}

	return output, nil

}

func main() {
	dagger.Serve(Kapp{})
}

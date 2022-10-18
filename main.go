package main

import (
	"fmt"

	"go.dagger.io/dagger/sdk/go/dagger"
	"go.dagger.io/dagger/sdk/go/dagger/api"
)

type Kapp struct {
}

func (Kapp) LoadCredentials(ctx dagger.Context, kubeconfig string) (api.SecretID, error) {

	client, err := dagger.Connect(ctx)
	if err != nil {
		return "", err
	}

	core := client.Core()

	secretId, err := core.Host().Workdir().Read().Directory(".").File(kubeconfig).Secret().ID(ctx)
	if err != nil {
		return "", err
	}

	return secretId, nil
}

func (Kapp) Deploy(ctx dagger.Context, app string, directory string, namespace string, url string, file string, kubeconfig api.SecretID) (string, error) {

	if app == "" {
		return "", fmt.Errorf("app name is required")
	}

	if namespace == "" {
		namespace = "default"
	}

	execOpts := api.ContainerExecOpts{
		Args: []string{"kapp", "deploy", "-y", "-n", namespace, "-a", app, "-f"},
	}

	client, err := dagger.Connect(ctx)
	if err != nil {
		return "", err
	}

	core := client.Core()

	fileId, err := core.Host().Workdir().Read().Directory(directory).File(file).ID(ctx)
	if err != nil {
		return "", err
	}

	containerId, err := core.Container().From("ghcr.io/vmware-tanzu/carvel-docker-image").WithMountedSecret("/root/.kube/config", kubeconfig).ID(ctx)
	if err != nil {
		return "", err
	}

	containerOpts := api.ContainerOpts{
		ID: containerId,
	}

	var output string
	if url == "" {
		execOpts.Args = append(execOpts.Args, "/source")
		output, err = core.Container(containerOpts).WithMountedFile("/source", fileId).Exec(execOpts).Stdout().Contents(ctx)
	} else {
		execOpts.Args = append(execOpts.Args, url)
		output, err = core.Container(containerOpts).Exec(execOpts).Stdout().Contents(ctx)
	}

	if err != nil {
		return "", err
	}

	return output, nil

}

func main() {
	dagger.Serve(Kapp{})
}

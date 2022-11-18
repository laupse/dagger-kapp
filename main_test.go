package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"dagger.io/dagger"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

func init() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Init : Buildkit started or already present")

	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	network, _ := cli.NetworkInspect(ctx, "kind", types.NetworkInspectOptions{})
	if err != nil {
		panic(err)
	}

	if len(network.Containers) < 2 {
		if err := cli.NetworkConnect(ctx, "kind", "dagger-buildkitd", nil); err != nil {
			panic(err)
		}
		fmt.Println("Init : Buildkit connected to kind network")
	}

	kindContainer, err := cli.ContainerInspect(ctx, "dagger-kapp-control-plane")
	if err != nil {
		panic(err)
	}

	kindIp := kindContainer.NetworkSettings.Networks["kind"].IPAddress
	kind := client.Host().Workdir().File("kind.yaml")

	_, err = client.Container().
		From("alpine:3.16").
		WithMountedFile("/kind.yaml", kind).
		Exec(dagger.ContainerExecOpts{
			Args: []string{"sed", "s/dagger-kapp-control-plane/" + kindIp + "/g", "kind.yaml"},
		}).Stdout().Export(ctx, "kind.ip.yaml")

	if err != nil {
		panic(err)
	}
	fmt.Println("Init : special kubeconfig for kind created")
}

func TestDeployWithFolder(t *testing.T) {

	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Error(err)
	}

	client.Host().Workdir().LoadProject("dagger.json")
	req := &dagger.Request{
		Query: `
		query MyQuery {
			kapp {
				deploy(
					app: "nginx-folder"
					directory: "./manifests"
					kubeconfig: "./kind.ip.yaml"
					namespace: "nginx"
					url: ""		
				)
			}
		}
		`,
	}
	var resp *dagger.Response

	err = client.Do(ctx, req, resp)
	if err != nil {
		t.Error(err)
	}

}

func TestDeployWithURL(t *testing.T) {

	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Error(err)
	}

	client.Host().Workdir().LoadProject("dagger.json")
	req := &dagger.Request{
		Query: `
		query MyQuery {
			kapp {
				deploy(
					app: "nginx-url"
					directory: ""
					kubeconfig: "./kind.ip.yaml"
					namespace: "nginx"
					url: "https://raw.githubusercontent.com/laupse/dagger-kapp/main/manifests/deploy.yaml"		
				)
			}
		}
		`,
	}
	var resp *dagger.Response

	err = client.Do(ctx, req, resp)
	if err != nil {
		t.Error(err)
	}

}

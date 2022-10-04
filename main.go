package main

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"go.dagger.io/dagger/sdk/go/dagger"
)

func (r *kapp) loadCredentials(ctx context.Context, fs dagger.FSID, kc string) (dagger.SecretID, error) {

	client, err := dagger.Client(ctx)
	if err != nil {
		return "", err
	}

	kubeConfigSecretId, err := loadCredentials(ctx, client, fs, kc)
	if err != nil {
		return "", err
	}
	return kubeConfigSecretId, nil

}

func loadCredentials(ctx context.Context, client graphql.Client, fs dagger.FSID, kubeConfig string) (dagger.SecretID, error) {
	req := &graphql.Request{
		Query: `
		query($fsid: FSID!, $kubeConfig: String!) {
			core {
				filesystem(id: $fsid) {
					file(path:  $kubeConfig)
				}
			}
		}
		`,
		Variables: map[string]any{
			"fsid":       fs,
			"kubeConfig": kubeConfig,
		},
	}
	respReadKc := struct {
		Core struct {
			Filesystem struct {
				File string
			}
		}
	}{}
	err := client.MakeRequest(ctx, req, &graphql.Response{Data: &respReadKc})
	if err != nil {
		return "", err
	}

	req = &graphql.Request{
		Query: `
		query($kubeConfig: String!){
			core {
			  	addSecret(plaintext: $kubeConfig)
			}
		}
		`,
		Variables: map[string]any{
			"kubeConfig": respReadKc.Core.Filesystem.File,
		},
	}
	respAddSecret := struct {
		Core struct {
			AddSecret dagger.SecretID
		}
	}{}
	err = client.MakeRequest(ctx, req, &graphql.Response{Data: &respAddSecret})
	if err != nil {
		return "", err
	}

	return respAddSecret.Core.AddSecret, nil
}

func (r *kapp) deploy(ctx context.Context, app *string, namespace *string, kubeConfig dagger.SecretID, url *string, file *string, fs dagger.FSID) (string, error) {

	if *app == "" {
		return "", fmt.Errorf("app name is required")
	}
	kappArgs := []string{"kapp", "deploy", "-y", "-n", *namespace, "-a", *app}

	if *url == "" {
		kappArgs = append(kappArgs, []string{"-f", "/source"}...)
	} else {
		kappArgs = append(kappArgs, []string{"-f", *url}...)
	}

	client, err := dagger.Client(ctx)
	if err != nil {
		return "", err
	}

	imageId, err := prepareImageWithCredentials(ctx, client, kubeConfig)
	if err != nil {
		return "", err
	}

	stdout, err := deploy(ctx, client, fs, imageId, *file, kappArgs)
	if err != nil {
		return "", err
	}

	return stdout, nil

}

func prepareImageWithCredentials(ctx context.Context, client graphql.Client, kubeConfig dagger.SecretID) (dagger.FSID, error) {
	req := &graphql.Request{
		Query: `
			query ($kubeConfig: SecretID!) {
				core {
					secret(id: $kubeConfig) 
				}
			}
		`,
		Variables: map[string]any{
			"kubeConfig": kubeConfig,
		},
	}
	respGetSecret := struct {
		Core struct {
			Secret string
		}
	}{}
	err := client.MakeRequest(ctx, req, &graphql.Response{Data: &respGetSecret})
	if err != nil {
		return "", err
	}

	req = &graphql.Request{
		Query: `
		query ($contents: String!) {
			core {
				image(ref: "ghcr.io/vmware-tanzu/carvel-docker-image") {
					exec(input: {args: ["mkdir", "-p", "/root/.kube"]}) {
						fs {
							writeFile(path: "/root/.kube/config", contents: $contents, permissions:"0400"){
								id
							}
						}
					}
				}
			}
		}
		`,
		Variables: map[string]any{
			"kubeConfig": kubeConfig,
			"contents":   respGetSecret.Core.Secret,
		},
	}
	respWriteKc := struct {
		Core struct {
			Image struct {
				Exec struct {
					Fs struct {
						WriteFile struct {
							ID dagger.FSID
						}
					}
				}
			}
		}
	}{}
	err = client.MakeRequest(ctx, req, &graphql.Response{Data: &respWriteKc})
	if err != nil {
		return "", err
	}

	return respWriteKc.Core.Image.Exec.Fs.WriteFile.ID, nil
}

func deploy(ctx context.Context, client graphql.Client, fsid dagger.FSID, image dagger.FSID, file string, args []string) (string, error) {
	req := &graphql.Request{
		Query: `
			query ($fsid: FSID!, $image: FSID!, $args: [String!]!, $file: String!) {
				core {
					filesystem(id: $image) {
						copy(from: $fsid, include: [$file], destPath: "/source"){
							exec(
								input:{
									args: $args,
								}) {
									stdout
							}
						}
					}
				}
			}
		`,
		Variables: map[string]any{
			"fsid":  fsid,
			"image": image,
			"args":  args,
			"file":  file,
		},
	}
	resp := struct {
		Core struct {
			Filesystem struct {
				Exec struct {
					Stdout string
				}
			}
		}
	}{}
	err := client.MakeRequest(ctx, req, &graphql.Response{Data: &resp})
	if err != nil {
		return "", err
	}

	return resp.Core.Filesystem.Exec.Stdout, nil
}

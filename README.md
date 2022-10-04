# dagger-kapp
A dagger extension for kapp operations

## Supported Commands
- deploy

## Include this in your cloak.yaml
```yaml
  - git:
      remote: https://github.com/laupse/dagger-kapp.git
      ref: main
      path: cloak.yaml
```

## Example
```gql
query LoadCred($fs: FSID!) {
  kapp {
    loadCredentials(fs: $fs, kc: "kubeconfig")
  }
}
```

```gql
query KappDeployKind($fs: FSID!, $kubeConfig: SecretID!) {
  kapp {
    deploy(fs: $fs, kubeConfig: $kubeConfig, app: "my-app", file: "deploy.yaml")
  }
}
```

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
query LoadCred() {
  kapp {
    loadCredentials(kubeconfig: "kubeconfig")
  }
}
```

```gql
query Deploy($kubeConfig: SecretID!) {
  kapp {
    deploy(app: "nginx", directory: ".", file: "deploy.yaml", kubeconfig: $kubeConfig,namespace: "default", url: "")
  }
}
```

# docker-credential-sync
Automatically sync docker credential between from a source namespace to other namespaces

## Quick Start

The operator is looking for a DockerCredentialSync resource in any namespace to read its configuration.

Assuming a user want to sync docker secret from `dev` namespace to other `dev-` namespaces

```
apiVersion: dax.vo/v1
kind: DockerCredentialSync
metadata:
  labels:
    app.kubernetes.io/name: docker-credential-sync
    app.kubernetes.io/managed-by: kustomize
  name: dockercredentialsync-sample
spec:
  sourceNamespace: dev
  sourceSecretName: registrypullsecret
  targetNamespacePrefix: "dev-"
  refreshIntervalSeconds: 3
```
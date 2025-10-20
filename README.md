# IPAM-API

# Setup

## Prerequisites
- go version v1.25.3+
- docker version 4.45.0+.
- MongoDB Compass 1.47.+

## Build Application
```sh
docker compose up --build
```

## Pre-populate Netbox with required configuration
```sh
python3 ./hack/netbox-scripts/setup_netbox_prefixes
```

# Operating

## MongoDB Compass

**Connection String:**`mongodb://admin:secretpassword@localhost:27017/?authSource=admin`

## Netbox

**URL:** ``http://localhost:8000``

**Username:** ``admin`` & **Password:** ``admin``

## Shell to IPAM-API

```sh
docker exec -it ipam-api bash
```

## IPAM-CLI

Executable within container IPAM-API.

```sh
./ipam-cli --help
```

# ArgoCD
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ipam-api
  namespace: argocd
spec:
  project: default
  source:
    path: .
    repoURL: oci://ncr.sky.nhn.no/ghcr/vitistack/helm/ipam-api
    targetRevision: 1.*
    helm:
      valueFiles:
          - values.prod.yaml
      parameters:
      - name: secrets.mongodb
        value: jsdfjkvnu253228asdfce385732sdfghwerh3545  
      - name: secrets.netbox
        value: yurmvld361sdfm4imdnej35ngotvm3583adkvnh3
      - name: secrets.splunk
        value: d5e33e86-3371-4d8c-b7a1-645c563f722e
      - name: encryption.encKey
        value: 86728dkfnhdj3744
      - name: encryption.encIv
        value: sdfkji4nfnkser45
      - name: backup.failedJobsHistoryLimit
        value: "6"
      - name: backup.sshKey
        value: |
          -----BEGIN OPENSSH PRIVATE KEY-----
          -----END OPENSSH PRIVATE KEY-----
  destination:
    server: "https://kubernetes.default.svc"
    namespace: ipam-system
  syncPolicy:
      automated:
          selfHeal: true
          prune: true
      syncOptions:
      - CreateNamespace=true
```
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

**URL** ``http://localhost:8000``

**Username:** ``admin`` & **Password:** ``admin``

## Shell to IPAM-API

``docker exec -it ipam-api bash``

## IPAM-CLI

Executable within container IPAM-API.

``./ipam-cli --help``
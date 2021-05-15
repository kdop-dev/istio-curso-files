# go-generic-service

## Variables

```bash
export SCHED_CALL_INTERVAL=5
export SCHED_CALL_URL_LST="http://httpbin.org/ip,http://httpbin.org/user-agent"
export SPLIT_CALL_URL_LST="http://httpbin.org/headers,http://httpbin.org/ip"
export APP="Test"
export VERSION="v1"
```

## Services

```bash
# Health
curl http://localhost:8000/healthz

# Split
curl http://localhost:8000/s

# Response code
curl http://localhost:8000/r?code=300&delay=1
```

## Running

```bash
# Create net
docker network create my-net

# backend
docker run -d --rm \
--hostname backend \
--network my-net \
--name backend \
-e SCHED_CALL_URL_LST="http://backend:8000/r?code=200&delay=0" \
-e SCHED_CALL_INTERVAL=10 \
-e SPLIT_CALL_URL_LST="http://backend:8000/r?code=200&delay=0" \
-e APP=backend \
-e VERSION=v1 \
kdop/go-generic-service:0.0.1

# front-end
docker run -d --rm \
    --network my-net \
    --hostname front-end \
    --name front-end \
    -e SCHED_CALL_URL_LST="http://front-end:8000/s" \
    -e SCHED_CALL_INTERVAL=10 \
    -e SPLIT_CALL_URL_LST="http://backend:8000/r?code=200&delay=1" \
    -e APP=front-end \
    -e VERSION=v1 \
    kdop/go-generic-service:0.0.1
```

## Helm

```bash
az login

az account set --subscription 73c8b399-a4ed-44b8-af3f-081a94ca1f65

az aks start --name kdop-learn --resource-group kdop-dev

az aks get-credentials --resource-group kdop-dev --name kdop-learn

az aks stop --name kdop-learn --resource-group kdop-dev
```

## Upgrade AKS

```bash
az aks get-upgrades --resource-group kdop-dev --name kdop-learn --output table

# Update max surge for an existing node pool - not configured
az aks nodepool update -n agentpool -g kdop-dev --cluster-name kdop-learn --max-surge 5

az aks upgrade \
    --resource-group kdop-dev \
    --name kdop-learn  \
    --kubernetes-version 1.20.5
```

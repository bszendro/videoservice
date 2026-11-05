
# Expose Prometheus metrics from service

This recipe shows how the metrics exposed by the service can be scraped by Prometheus and displayed in Grafana deployed inside a Kubernetes cluster.

## Start the cluster

```bash
$ kind create cluster --config kind.yaml
```

## Deploy the Prometheus stack

```bash
$ kubectl create namespace monitoring
$ helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
$ helm install prom prometheus-community/kube-prometheus-stack --version 92.17.1 -n monitoring
# Wait until all pods become ready
$ helm test prom -n monitoring
# Get the login password for Grafana
$ kubectl get secret prom-grafana -n monitoring -o jsonpath="{.data.admin-password}" | base64 --decode; echo
# Make Grafana accessible over localhost:3000
$ kubectl port-forward service/prom-grafana -n monitoring 3000:80
```

* Open [http://localhost:3000/](http://localhost:3000/)
* Login with admin and the password read from the secret previously

## Build the image

```bash
$ docker build -t ghcr.io/bszendro/videoservice/videoservice:1.0.0 .
# Login the first time
$ docker login ghcr.io -u bszendro
$ docker push ghcr.io/bszendro/videoservice/videoservice:1.0.0
```

## Deploy the image using Helm

> **_NOTE:_** When using a private repository on GHCR, create a Personal access token (classic) on GitHub with `write:packages` permissions, then store the secret in k8s. (Note that this secret is not encrypted in the DB by default.)
> ```bash
> $ kubectl create secret docker-registry ghcr-access-key --docker-server=ghcr.io --docker-username=bszendro --docker-password=<access-key>
>
> $ kubectl get secret ghcr-access-key
> ```
>

```bash
$ helm install vidserv ./vidserv --version 1.0.0
# Wait until all pods become ready
$ helm test vidserv
```

The metrics exposed by the service pods should be visible in Grafana's Explore view now.
The way this works is that the Prometheus stack operator detects the `PodMonitor` resource created by Helm when deploying the service. The `PodMonitor` resource in the chart has a `release` label that matches the selector value in the `Prometheus` resource:

```bash
# Check what podmonitors will this Prometheus instance use
$ kubectl get prometheus prom-kube-prometheus-stack-prometheus -n monitoring -o json | jq '.spec.podMonitorSelector.matchLabels'
{
  "release": "prom"
}
```

## What are the scrape targets visible to Prometheus?

```bash
$ kubectl port-forward service/prom-kube-prometheus-stack-prometheus -n monitoring 9090
```

* Open [http://localhost:9090/targets](http://localhost:9090/targets)

## Test network connectivity from another pod inside the cluster

```bash
$ kubectl apply -f netshoot-client-pod.yaml
$ kubectl exec pod/netshoot-client -- curl http://10.244.1.8:8000/hello
```

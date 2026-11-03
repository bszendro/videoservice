
## Start the cluster

```bash
kind create cluster --config kind.yaml
```

## Build the image

```bash
docker build -t ghcr.io/bszendro/videoservice/videoservice:latest .
docker push ghcr.io/bszendro/videoservice/videoservice:latest
```

## Deploy the image

> **_NOTE:_** When using a private repository, you need to create the secret in k8s first:
> ```bash
> $ kubectl create secret docker-registry ghcr-access-key \
>       --docker-server=ghcr.io \
>       --docker-username=bszendro \
>       --docker-password=<access-key>
>
> $ kubectl get secret ghcr-access-key
> ```
>

```bash
kubectl apply -f videoservice-deployment.yaml
kubectl apply -f videoservice-service.yaml
```

## Deploy Prometheus using Helm

```bash
helm install myprom oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack:92.17.1
```

## Connect to Grafana

```bash
kubectl port-forward service/myprom-grafana 3000:80
```

Get the password for `admin`:
```bash
kubectl get secret myprom-grafana -o jsonpath="{.data.admin-password}" | base64 --decode; echo
```

Open [http://localhost:3000/](http://localhost:3000/)

## Add the service to Prometheus targets

```bash
kubectl apply -f videoservice-smon.yaml
```

How this works: The `Prometheus` CRD has `release=myprom` in `spec.serviceMonitorSelector.matchLabels` which will make the operator pick up and apply `videoservice-smon.yaml` that has this label.

After this step the exposed metrics of the service should be visible in Grafana.

## Test network connectivity from another pod inside the cluster

```bash
kubectl apply -f netshoot-client-pod.yaml
kubectl exec pod/netshoot-client -- curl http://10.244.1.8:8000/hello
```

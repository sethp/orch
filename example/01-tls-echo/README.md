# example

```
make docker-run
```

```
curl -v --cacert ./server.crt https://localhost:8443/hello-world
```

## the manual way

```
make docker-build kind-load

kubectl create secret tls tls-echo --cert=./server.crt --key=./server.key --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f ./manifest/manifest.yaml

kubectl port-forward svc/tls-echo 8443:443
```

```
curl -v --cacert ./example/server.crt https://localhost:8443/hello-world
```

## with orch

```
make docker-build kind-load

go run ./cmd/orch generate | kubectl apply -f -

kubectl port-forward svc/tls-echo 8443:443
```

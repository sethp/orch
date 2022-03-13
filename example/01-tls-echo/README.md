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

kubectl create secret tls tls-echo --cert=./server.crt --key=./server.key
kubectl apply -f ./manifest/manifest.yaml

kubectl port-forward svc/tls-echo 8443:443
```

```
curl -v --cacert ./example/server.crt https://localhost:8443/hello-world
```

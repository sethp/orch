# example: metallb-kind

To install & configure metallb in `kind`:

```
go run ./cmd/orch generate | kubectl apply -f -
```

Based on: https://kind.sigs.k8s.io/docs/user/loadbalancer/

# orch

a tool for building up toward full orchestration with kubernetes

## rough sketch

1. starts with yaml or spooky yaml at a distance (i.e. helm chart), but, "I need to run just a little bit of logic", and, I want to do just a little bit of higher-level config ("features: [x, y, z]")
2. `helm template > manifest/chart.yaml` or `kustomize build > manifest/app.yaml`
3. `orch init`
4. <write a little bit of go> (naml may be helpful here, either as a reference or directly)
   - examples of how/when to pull chunks of a resource or all of a resource into go structs?
5. iterate on config, lifecycle logic
   - is this the bridge? `orch` sets up an example project that makes it easy(ish?) to write kube-native code?
6. cli pivots to operator reconciling CRD

## some possible examples

1. simple app
   1. with tls (cert-manager or nah)
2. cli for setting up metallb on kind
3. prometheus operator something something
   1. what are the most commonly twiddled options, I wonder?

maybe cert-manager-less webhooks?

## important details

1. server-side apply
2. what happens when two clis run concurrently? at different versions?

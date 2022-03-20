

.PHONY: manifest
manifest: manifest/metallb.yaml

# Not phony
manifest/metallb.yaml: Makefile
	curl -fsSo $@ https://raw.githubusercontent.com/metallb/metallb/c7aef7814e2be3e78f68864825c7609b000109f2/manifests/metallb.yaml


.PHONY: example-app
example-app:
	kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/usage.yaml


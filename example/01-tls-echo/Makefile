
export DOCKER_BUILDKIT=1
IMG ?= tls-echo

STEP ?= $(if $(shell command -v step-cli),step-cli,step)

.PHONY: docker-build
docker-build:
	docker build . -t $(IMG)

.PHONY: kind-load
kind-load:
	kind load docker-image $(IMG)

.PHONY: docker-run
docker-run: docker-build server.crt server.key
	docker run --rm -p 8443:8443 \
		-v `pwd`/server.crt:/etc/tls/tls.crt \
		-v `pwd`/server.key:/etc/tls/tls.key \
		$(IMG)

# not phony
%.crt %.key: Makefile # depend on Makefile to re-run when the recipe changes (as well as any other change, unfortuantely)
# this recipe is fragile in that it requires `docker-run` to depend on exactly `server.crt` and `server.key`,
# but we have to use a pattern recipe here to get make to portably recognize that it should only regenerate
#
# in gnu make 4.3 we could use `server.crt server.key &: ...` to remove this fragility
	$(STEP) certificate create \
		--profile self-signed \
		--subtle --no-password --insecure \
		--san localhost --san 127.0.0.1 \
		--force \
		CN=test-cert \
		server.crt server.key
	chmod o+r server.crt server.key

.NOTPARALLEL:

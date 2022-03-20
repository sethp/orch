BIN_DIR := ./hack/bin/$(shell go env GOOS)_$(shell go env GOARCH)

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

GOLANGCI_LINT_VERSION ?= v1.45.0

.PHONY: lint
lint: lint.go

.PHONY: lint.go
lint.go: $(GOLANGCI_LINT) $(GOLANGCI_LINT).$(GOLANGCI_LINT_VERSION).stamp
	$(GOLANGCI_LINT) run

# not phony
$(GOLANGCI_LINT) $(GOLANGCI_LINT).$(GOLANGCI_LINT_VERSION).stamp :
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(dir $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION)
	rm -f $(GOLANGCI_LINT).*.stamp
	touch $(GOLANGCI_LINT).$(GOLANGCI_LINT_VERSION).stamp

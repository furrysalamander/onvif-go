GOCMD     := go
GOFLAGS   :=
PACKAGES  := ./...
BINDIR    := bin

GO_GENERATE := $(GOCMD) generate $(GOFLAGS) $(PACKAGES)
GO_VET     := $(GOCMD) vet $(GOFLAGS) $(PACKAGES)
GO_BUILD   := $(GOCMD) build $(GOFLAGS) $(PACKAGES)
GO_TEST    := $(GOCMD) test $(GOFLAGS) $(PACKAGES)

GOLANGCI_LINT_VERSION := v2.12.2
GOLANGCI_LINT_BIN    := $(BINDIR)/golangci-lint

.PHONY: all
all: generate lint vet build test

.PHONY: generate
generate:
	$(GO_GENERATE)

.PHONY: vet
vet:
	$(GO_VET)

.PHONY: build
build:
	$(GO_BUILD)

.PHONY: test
test:
	$(GO_TEST)

.PHONY: golden-update
golden-update:
	$(GOCMD) test $(GOFLAGS) $(PACKAGES) -run TestGolden -update

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run ./...

# Ensures the generator output matches what's committed; fails on drift.
.PHONY: check-gen
check-git-clean:
	@if ! git diff --quiet --exit-code -- onvif/ ; then \
		echo "ERROR: generated code under onvif/ is out of date."; \
		echo "Run 'go generate ./...' and commit the result."; \
		git --no-pager diff --stat -- onvif/ ; \
		exit 1 ; \
	fi

.PHONY: check-gen
check-gen: generate check-git-clean

$(GOLANGCI_LINT_BIN):
	@mkdir -p $(BINDIR)
	@GOBIN=$(CURDIR)/$(BINDIR) $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: tools
tools: $(GOLANGCI_LINT_BIN)

.PHONY: clean
clean:
	rm -rf $(BINDIR) coverage.txt coverage.html
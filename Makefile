all: help

-include rules.mk

boots: cmd/boots/boots ## Compile boots for host OS and Architecture

ipxe: $(generated_ipxe_files) ## Build all iPXE binaries 

crosscompile: $(crossbinaries) ## Compile boots for all architectures
	
gen: $(generated_files) ## Generate go generate'd files

image: cmd/boots/boots-linux-amd64 ## Build docker image
	docker build -t boots .

stack-run: cmd/boots/boots-linux-amd64 ## Run the Tinkerbell stack
	cd deploy/stack; docker-compose up --build -d

stack-remove: ## Remove a running Tinkerbell stack
	cd deploy/stack; docker-compose down -v --remove-orphans

test: gen ipxe ## Run go test
	CGO_ENABLED=1 go test -race -coverprofile=coverage.txt -covermode=atomic ${TEST_ARGS} ./...

test-ipxe: ipxe/tests ## Run iPXE feature tests

coverage: test ## Show test coverage
	go tool cover -func=coverage.txt

vet: ## Run go vet
	go vet ./...

goimports: ## Run goimports
	@echo be sure goimports is installed
	goimports -w .

golangci-lint: ## Run golangci-lint
	@echo be sure golangci-lint is installed: https://golangci-lint.run/usage/install/
	golangci-lint run

validate-local: vet coverage goimports golangci-lint ## Runs all the same validations and tests that run in CI

help: ## Print this help
	@grep --no-filename -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sed 's/:.*##/·/' | sort | column -ts '·' -c 120

GO_INSTALL = ./scripts/go_install.sh
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)

CONTROLLER_GEN_VER := v0.2.9
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER)

ENVSUBST_BIN := envsubst
ENVSUBST := $(TOOLS_BIN_DIR)/$(ENVSUBST_BIN)-drone

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(CONTROLLER_GEN): ## Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) sigs.k8s.io/controller-tools/cmd/controller-gen $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)


.PHONY: generate
generate: ## Generate code, manifests etc.
	$(MAKE) generate-go
	$(MAKE) generate-manifests

.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) # Generate Go code.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate/boilerplate.generatego.txt" paths="./k8s/..."

.PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) # Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./k8s/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook

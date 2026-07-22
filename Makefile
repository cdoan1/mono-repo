MODULES = api sdk tools cli

BIN_DIR = bin

HYPERSHIFT_IMPORT_PATH = github.com/openshift/hypershift/api/hypershift/v1beta1
HYPERSHIFT_TYPES = HostedClusterSpec,NodePoolSpec

.PHONY: all
all: build test vet ## Build, test, and vet all modules

.PHONY: build
build: ## Build all modules
	@mkdir -p $(BIN_DIR)
	@for m in $(MODULES); do \
		echo "=== Building $$m ==="; \
		if [ "$$m" = "cli" ]; then \
			cd $$m && go build -o ../$(BIN_DIR)/hyperfleetctl ./ && cd ..; \
		else \
			cd $$m && go build ./... && cd ..; \
		fi \
	done

.PHONY: test
test: ## Test all modules
	@for m in $(MODULES); do \
		echo "=== Testing $$m ==="; \
		cd $$m && go test -v -race ./... && cd ..; \
	done

.PHONY: fmt
fmt: ## Format all Go source files
	gofmt -w -s .

.PHONY: vet
vet: ## Vet all modules
	@for m in $(MODULES); do \
		echo "=== Vetting $$m ==="; \
		cd $$m && go vet ./... && cd ..; \
	done

.PHONY: tidy
tidy: ## Run go mod tidy on all modules
	@for m in $(MODULES); do \
		echo "=== Tidying $$m ==="; \
		cd $$m && go mod tidy && cd ..; \
	done

.PHONY: lint
lint: ## Run golangci-lint on all modules
	@for m in $(MODULES); do \
		echo "=== Linting $$m ==="; \
		cd $$m && golangci-lint run ./... && cd ..; \
	done

# === Build tools ===

.PHONY: build-tools
build-tools: ## Build all codegen CLI tools into bin/
	@mkdir -p $(BIN_DIR)
	cd tools && go build -o ../$(BIN_DIR)/passthrough-gen ./cmd/passthrough-gen
	cd tools && go build -o ../$(BIN_DIR)/marker-scanner ./cmd/marker-scanner
	cd tools && go build -o ../$(BIN_DIR)/openapi-gen ./cmd/openapi-gen
	cd tools && go build -o ../$(BIN_DIR)/conversion-gen ./cmd/conversion-gen
	cd tools && go build -o ../$(BIN_DIR)/crd-variants ./cmd/crd-variants
	cd tools && go build -o ../$(BIN_DIR)/featuregate-info ./cmd/featuregate-info
	cd tools && go build -o ../$(BIN_DIR)/verify-configuration ./cmd/verify-configuration

.PHONY: build-cli
build-cli: ## Build the CLI binary into bin/
	@mkdir -p $(BIN_DIR)
	cd cli && go build -o ../$(BIN_DIR)/hyperfleetctl ./

# === Code generation (placeholder targets) ===

.PHONY: generate
generate: generate-registry generate-passthrough generate-openapi generate-conversion ## Run all generators

.PHONY: generate-registry
generate-registry: build-tools ## Generate field metadata registry from Go markers
	$(BIN_DIR)/marker-scanner --input-dirs=api/v1alpha1 --output-file=sdk/registry/field_metadata.go

.PHONY: generate-registry-verbose
generate-registry-verbose: build-tools ## Generate field metadata registry (verbose output)
	$(BIN_DIR)/marker-scanner --input-dirs=api/v1alpha1 --output-file=sdk/registry/field_metadata.go --verbose

.PHONY: generate-passthrough
generate-passthrough: build-tools ## Generate passthrough types from HyperShift
	$(BIN_DIR)/passthrough-gen --import-path=$(HYPERSHIFT_IMPORT_PATH) --types=$(HYPERSHIFT_TYPES) --output-dir=api/v1alpha1 --package=v1alpha1

.PHONY: generate-openapi
generate-openapi: build-tools ## Generate OpenAPI JSON schema
	$(BIN_DIR)/openapi-gen --input-dirs=api/v1alpha1 --output-file=openapi/openapi.json

.PHONY: generate-conversion
generate-conversion: build-tools ## Generate REST types and conversion functions
	$(BIN_DIR)/conversion-gen --api-version=v1alpha1 --crd-package=github.com/cdoan1/mono-repo/api/v1alpha1 --input-dirs=api/v1alpha1 --output-dir=tools/pkg/conversion/v1alpha1

# === Workspace ===

.PHONY: workspace-sync
workspace-sync: ## Sync go.work file
	go work sync

.PHONY: get-hypershift-version
get-hypershift-version: ## Show current HyperShift version in go.mod
	@PSEUDO_VERSION=$$(grep "github.com/openshift/hypershift/api" api/go.mod | awk '{print $$2}'); \
	COMMIT=$$(echo $$PSEUDO_VERSION | rev | cut -d'-' -f1 | rev); \
	echo "Current HyperShift in go.mod:"; \
	echo "  Pseudo-version: $$PSEUDO_VERSION"; \
	echo "  Commit: $$COMMIT"; \
	TAG=$$(curl -s https://api.github.com/repos/openshift/hypershift/tags | jq -r ".[] | select(.commit.sha | startswith(\"$$COMMIT\")) | .name" | head -1); \
	if [ -z "$$TAG" ]; then \
		echo "  Tag: (no tag found - using commit)"; \
	else \
		echo "  Tag: $$TAG"; \
	fi

# === Verify ===

.PHONY: verify
verify: fmt vet test ## Run full verification suite

.PHONY: ci
ci: build test vet ## CI pipeline

# === Clean ===

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out

.PHONY: clean-generated
clean-generated: ## Remove generated files
	rm -f openapi/openapi.json
	rm -f config/crd/bases/*.yaml
	rm -f config/crd/variants/*.yaml

# === Help ===

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

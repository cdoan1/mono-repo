MODULES = api sdk tools cli

BIN_DIR = bin

HYPERSHIFT_IMPORT_PATH = github.com/openshift/hypershift/api/hypershift/v1beta1
HYPERSHIFT_TYPES = HostedClusterSpec,NodePoolSpec

.PHONY: all
all: build test vet ## Build, test, and vet all modules

.PHONY: build
build: ## Build all modules
	@for m in $(MODULES); do \
		echo "=== Building $$m ==="; \
		cd $$m && go build ./... && cd ..; \
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
	cd cli && go build -o ../$(BIN_DIR)/hyperfleet ./

# === Code generation (placeholder targets) ===

.PHONY: generate
generate: ## Run all generators (placeholder)
	@echo "TODO: generate-registry generate-passthrough manifests generate-openapi generate-conversion"

.PHONY: generate-registry
generate-registry: ## Generate field metadata registry from Go markers
	@echo "TODO: bin/marker-scanner --input-dirs=api/v1alpha1 --output-file=sdk/registry/field_metadata.go"

.PHONY: generate-passthrough
generate-passthrough: ## Generate passthrough types from HyperShift
	@echo "TODO: bin/passthrough-gen --import-path=$(HYPERSHIFT_IMPORT_PATH) --types=$(HYPERSHIFT_TYPES) --output-dir=api/v1alpha1 --package=v1alpha1"

.PHONY: generate-openapi
generate-openapi: ## Generate OpenAPI JSON schema
	@echo "TODO: bin/openapi-gen --input-dirs=api/v1alpha1 --output-file=openapi/openapi.json"

.PHONY: generate-conversion
generate-conversion: ## Generate REST types and conversion functions
	@echo "TODO: bin/conversion-gen --api-version=v1alpha1 --crd-package=api/v1alpha1 --output-dir=tools/pkg/conversion"

# === Workspace ===

.PHONY: workspace-sync
workspace-sync: ## Sync go.work file
	go work sync

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

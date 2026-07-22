# mono-repo

Go mono-repo for the HyperFleet API — code generation tools, SDK, and CLI for bridging HyperShift CRDs to a Platform REST API.

## Modules

| Module | Path | Description |
|--------|------|-------------|
| **api** | `github.com/cdoan1/mono-repo/api` | CRD type definitions (Cluster, NodePool, Configuration, passthrough types) |
| **tools** | `github.com/cdoan1/mono-repo/tools` | Code generators: passthrough-gen, marker-scanner, openapi-gen, conversion-gen, crd-variants, featuregate-info, verify-configuration |
| **sdk** | `github.com/cdoan1/mono-repo/sdk` | Runtime libraries: field registry, feature gates, validation, API client |
| **cli** | `github.com/cdoan1/mono-repo/cli` | CLI leveraging the SDK and generated types |

## Quick Start

```bash
# Build all modules (cli binary goes to bin/hyperfleetctl)
make build

# Build codegen tools into bin/
make build-tools

# Run all generators
make generate

# Run tests (with race detector)
make test

# Full verification (fmt + vet + test)
make verify
```

## Code Generation

`make generate` runs all four generators in sequence:

| Target | Command | Output |
|--------|---------|--------|
| `generate-registry` | `marker-scanner` | `sdk/registry/field_metadata.go` + `.json` |
| `generate-passthrough` | `passthrough-gen` | `api/v1alpha1/hostedclusterspec.passthrough.go` |
| `generate-openapi` | `openapi-gen` | `openapi/openapi.json` |
| `generate-conversion` | `conversion-gen` | `tools/pkg/conversion/v1alpha1/` |

Individual generators can be run separately, e.g. `make generate-registry`.

Use `make generate-registry-verbose` to see the full field table (122 fields with write modes, feature gates, and visibility).

## Architecture

The project uses a [Go workspace](https://go.dev/doc/tutorial/workspaces) (`go.work`) to manage four modules. The API module has its own dependency graph (k8s.io, openshift) separate from the tools.

```
cli -> sdk -> api
tools -> sdk, api
```

### Codegen Pipeline

```
HyperShift CRD -> [passthrough-gen] -> HyperFleet CRD -> [openapi-gen] -> Platform API
                                            |
                                     [marker-scanner] -> field_metadata.go/json -> [validation]
                                            |
                                     [conversion-gen] -> REST types + conversion functions
```

### Three Control Markers

| Marker | Purpose | Values |
|--------|---------|--------|
| `+k8s:openapi-gen` | Visibility | `true` (visible) / `false` (hidden) |
| `+hyperfleet:write-mode` | Write access | `mutable` / `immutable` / `service-set` |
| `+openshift:enable:FeatureGate` | Feature gating | Gate name (e.g. `HyperFleetEtcdConfig`) |

### Field Registry Summary

122 fields scanned from `api/v1alpha1/`:
- **Write Modes**: 29 mutable, 6 immutable, 87 service-set
- **Visibility**: 44 visible, 78 hidden
- **Gating**: 8 feature-gated, 114 ungated

## Make Targets

```
make help              # Show all targets
make build             # Build all modules
make test              # Test all modules (with -race)
make fmt               # Format all Go source
make vet               # Vet all modules
make tidy              # go mod tidy all modules
make lint              # Run golangci-lint
make build-tools       # Build 7 codegen tools into bin/
make build-cli         # Build CLI into bin/hyperfleetctl
make generate          # Run all code generators
make generate-registry          # Generate field metadata registry
make generate-registry-verbose  # Generate registry with field table
make generate-passthrough       # Generate passthrough types from HyperShift
make generate-openapi           # Generate OpenAPI JSON schema
make generate-conversion        # Generate REST types and conversions
make get-hypershift-version     # Show pinned HyperShift version
make workspace-sync    # Sync go.work file
make verify            # fmt + vet + test
make ci                # build + test + vet
make clean             # Remove bin/ and coverage.out
make clean-generated   # Remove generated files
```

## HyperShift Dependency

Both `api/` and `tools/` pin the same HyperShift version. Check the current pin:

```bash
make get-hypershift-version
```

## Reference

Reproduced from [cdoan1/hyperfleet-api-codegen](https://github.com/cdoan1/hyperfleet-api-codegen).

---

## Generation Prompt

See [docs/p1.md](docs/p1.md) for the generation prompt and project requirements.

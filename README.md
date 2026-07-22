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
# Build all modules
make build

# Build codegen tools
make build-tools

# Run tests
make test

# Format and vet
make fmt
make vet

# Tidy all modules
make tidy

# Run all generators
make generate
```

## Architecture

The project uses a [Go workspace](https://go.dev/doc/tutorial/workspaces) (`go.work`) to manage four modules. The API module has its own dependency graph (k8s.io, openshift) separate from the tools.

```
HyperShift CRD → [passthrough-gen] → HyperFleet CRD → [openapi-gen] → Platform API
                                          ↓
                                   [marker-scanner] → field_metadata.json → [validation]
```

## Reference

Reproduced from [cdoan1/hyperfleet-api-codegen](https://github.com/cdoan1/hyperfleet-api-codegen).

---

## Generation Prompt

See [docs/p1.md](docs/p1.md) for the generation prompt and project requirements.

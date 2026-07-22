# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Go mono-repo that restructures `cdoan1/hyperfleet-api-codegen` into a multi-module layout. The project bridges HyperShift CRDs to a Platform REST API using marker-based code generation.

## Build Commands

```bash
make build           # Build all modules
make test            # Test all modules (with -race)
make fmt             # Format all Go source
make vet             # Vet all modules
make tidy            # go mod tidy all modules
make build-tools     # Build codegen tools into bin/
make build-cli       # Build CLI into bin/
make generate        # Run all code generators
make clean           # Remove bin/ and coverage.out
make help            # Show all targets
```

Single module operations:
```bash
cd api && go build ./...
cd api && go test -v -race ./...
```

## Architecture

Four Go modules tied together by `go.work`:

- **api/** (`github.com/cdoan1/mono-repo/api`) — CRD type definitions: Cluster, NodePool, Configuration, passthrough types. Has its own dependency graph (k8s.io, openshift/hypershift).
- **tools/** (`github.com/cdoan1/mono-repo/tools`) — Code generators (7 CLI tools under `tools/cmd/`) and codegen libraries (`tools/pkg/markers`, `tools/pkg/passthrough`, `tools/pkg/openapi`, `tools/pkg/conversion`).
- **sdk/** (`github.com/cdoan1/mono-repo/sdk`) — Runtime libraries: field registry types (`sdk/registry`), feature gate system (`sdk/featuregate`), request validation (`sdk/validation`), API client (`sdk/client`).
- **cli/** (`github.com/cdoan1/mono-repo/cli`) — CLI that leverages the SDK and generated types.

### Module Dependencies

```
cli → sdk → api
tools → sdk, api
```

The API module has no internal dependencies. The SDK depends only on the API module. The tools module depends on both API and SDK.

### Key Patterns

- **Three control markers**: Visibility (`+k8s:openapi-gen`), Write Mode (`+hyperfleet:write-mode={mutable|immutable|service-set}`), Feature Gate (`+openshift:enable:FeatureGate`).
- **Field registry as memory**: `sdk/registry/field_metadata.json` feeds runtime validation AND preserves developer-curated markers across passthrough regeneration on HyperShift bumps.
- **Two-boundary codegen**: HyperShift CRD → HyperFleet CRD (passthrough gen), then HyperFleet CRD → Platform API (OpenAPI + field metadata).

### Generated Outputs

| Output | Location | Generator |
|--------|----------|-----------|
| Passthrough types | `api/v1alpha1/hostedclusterspec.passthrough.go` | passthrough-gen |
| Field registry | `sdk/registry/field_metadata.go` | marker-scanner |
| OpenAPI schema | `openapi/openapi.json` | openapi-gen |
| CRD manifests | `config/crd/bases/` | controller-gen |
| CRD variants | `config/crd/variants/` | crd-variants |

## Style Rules

- Design documents go in `./docs/`.
- Update `README.md` as appropriate.
- Keep the generation prompt at the bottom of `README.md` so it can be edited and re-run.

## Testing

- Comprehensive unit tests for all components.
- Comprehensive e2e tests; e2e completion progress tracks overall mono-repo project progress.
- Run tests with race detector: `go test -v -race ./...`

## Reference

The original single-repo implementation is at `/Users/cdoan/workspace/src/github.com/cdoan1/hyperfleet-api-codegen/`.

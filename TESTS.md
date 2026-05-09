# screenjson-cli Tests

## Overview

This repository uses Go unit tests to validate the ScreenJSON data pipeline, format bridges, JSON codec behavior, model serialization, and schema validation.

The tests are designed to cover:

- canonical model behavior, including `Slugline.No` and `Element.SceneNo`
- format bridge conversion for Fountain, Celtx, FDX, FadeIn, and PDF pathways
- JSON codec encoding/decoding round trips
- pipeline builder/writer round trips
- schema validation logic and embedded schema structure

## Run the full test suite

From the repository root:

```bash
cd /Users/jason/Projects/screenjson-cli
go test ./...
```

This command runs all tests in every package under `internal/`.

## Run package-specific tests

Use package paths to run tests for a smaller scope.

- Fountain bridge tests:

```bash
go test ./internal/formats/fountain/bridge
```

- JSON codec tests:

```bash
go test ./internal/formats/json/codec
```

- Schema validator tests:

```bash
go test ./internal/schema
```

- Pipeline tests:

```bash
go test ./internal/pipeline
```

- Model tests:

```bash
go test ./internal/model
```

## Common test commands

- Run a single test by name:

```bash
go test ./internal/formats/fountain/bridge -run TestFountainEndToEndSceneNumber
```

- Run tests with verbose output:

```bash
go test ./... -v
```

- Disable result caching:

```bash
go test ./... -count=1
```

- Run tests with the race detector:

```bash
go test ./... -race
```

## Notes

- Some packages do not contain test files, so `go test` may report `[no test files]` for those packages.
- Tests currently focus on the ScreenJSON conversion and validation layers; format-specific helper packages may remain untested if they do not include dedicated test files.
- The repository is Go module based, so the root directory must contain the module file (`go.mod`) when running tests.

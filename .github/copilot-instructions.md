---
name: screenjson-cli
description: "Workspace instructions for screenjson-cli: A Go CLI tool for converting, validating, and exporting screenplays using the ScreenJSON format."
---

# screenjson-cli Workspace Instructions

## Project Overview

**screenjson-cli** is a command-line tool and REST API server for screenplay document processing. It converts between multiple screenplay formats (FDX, Fountain, FadeIn, PDF, Celtx) and the canonical ScreenJSON format, with support for validation, encryption, and REST API exposure.

**Key capabilities**:
- Convert screenplays between 7 formats
- Validate documents against JSON Schema
- AES-256 encryption/decryption
- REST API server (`/convert`, `/export`, `/validate` endpoints)
- Multi-format output (JSON/YAML)
- Database and blob store integrations (S3, Azure, MinIO)

## Build & Runtime

### Build Command
```bash
go build -o screenjson ./cmd/screenjson/
```

### Prerequisites
- **Poppler utility** (`pdftohtml`) for PDF import — set environment variable:
  ```bash
  # macOS Intel
  export SCREENJSON_PDFTOHTML=/usr/local/bin/pdftohtml
  
  # macOS Apple Silicon
  export SCREENJSON_PDFTOHTML=/opt/homebrew/bin/pdftohtml
  
  # Linux
  export SCREENJSON_PDFTOHTML=/usr/bin/pdftohtml
  ```
- **Docker Compose** (optional): Orchestrates screenjson + Gotenberg (PDF gen) + Tika (OCR)

### CLI Commands
```bash
screenjson convert -i input.fdx -o output.json
screenjson export -i screenplay.json -f pdf -o screenplay.pdf
screenjson validate -i screenplay.json
screenjson encrypt -i screenplay.json -k <key>
screenjson decrypt -i screenplay-encrypted.json -k <key>
screenjson serve --port 8080
screenjson formats
screenjson version
```

## Architecture

```
cmd/screenjson/
├── main.go                      # CLI dispatcher

internal/
├── app/app.go                   # Dependency injection container (lazy-init external clients)
├── model/                       # ✓ Canonical domain model (zero external imports)
│   ├── document.go              # Root Document with authors, characters, content
│   ├── elements.go              # Typed element union (action|character|dialogue|etc.)
│   ├── characters.go            # UUID-based character definitions
│   ├── scene.go                 # Scene heading (slugline) + body elements
│   ├── types.go                 # Shared types (Slugline, Text, Revision, etc.)
│   └── errors.go                # ParseError with non-fatal code + context
│
├── formats/                     # Format codec implementations
│   ├── {fdx,fountain,fadein,pdf,celtx}/
│   │   ├── codec/               # Decoder/Encoder (binary/text ↔ internal model)
│   │   ├── bridge/              # Format-specific model → canonical Document
│   │   └── model/               # Format-specific intermediate types
│   ├── json/codec/              # JSON codec (identity on Document)
│   └── yaml/codec/              # YAML codec (identity on Document)
│
├── pipeline/                    # Format registry & processing
│   ├── registry.go              # Format registry + capabilities mapping
│   ├── build_document.go        # Decoder dispatch + bridge transformation
│   ├── write_output.go          # Encoder dispatch
│   └── job.go                   # Async job status + queue integration
│
├── schema/                      # JSON Schema validation
│   ├── schema.go                # Validator wrapper
│   ├── schema.json              # Embedded canonical schema
│   └── validator.go             # santhosh-tekuri/jsonschema validation
│
├── config/config.go             # CLI flag → environment → defaults precedence
├── crypto/                      # AES-256 GCM encryption
│   └── aes.go, encrypt.go
├── external/                    # External service clients
│   ├── gotenberg/client.go      # PDF conversion
│   ├── llm/client.go            # LLM analysis (future)
│   └── tika/                    # OCR + document extraction
├── server/server.go             # REST API (handlers + routes)
├── stores/                      # Storage backends (interface + implementations)
│   ├── interface.go
│   └── {cassandra,chromadb,dynamodb,elasticsearch,mongodb,pinecone,redis,weaviate}/
└── queue/queue.go               # Job queue for async processing
```

## Design Patterns

### 1. **Model Package — Canonical Boundary**
The `model/` package:
- Zero external imports → pure domain logic
- All format codecs **must** bridge their output to `model.Document`
- Source of truth for screenplay structure (validated against `schema.json`)

**Key types**:
- `Document`: Root with authors, characters, revisions, scenes
- `Element`: Discriminated union with `Type` + typed fields (Text, Character, etc.)
- `Text`: Multi-language map: `{"en": "...", "fr": "..."}`
- `Character`: UUID-based (eliminates "JOHN" vs "JOHNNY" ambiguity)
- `ParseError`: Non-fatal errors with code + line/column + context

### 2. **Format Codec Triple Pattern**
Every format follows:
1. **Decoder** (`codec/decode.go`): Parse input → format-specific model
2. **Bridge** (`bridge/bridge.go`): Format model → `model.Document`
3. **Encoder** (`codec/encode.go`): `model.Document` → output bytes

**Examples**:
- **Fountain**: Regex-based element classification; extracts authors/characters during bridge
- **PDF**: Runs `pdftohtml -xml` subprocess; geometry-based margin clustering for element typing

### 3. **Format Registry**
Central `Registry` in `pipeline/registry.go`:
- Maps extensions (case-insensitive) → format handlers
- Stores capabilities: can decode, can encode, MIME types
- Auto-detects format from file extension or explicit `--format` flag

### 4. **Application Container** (`app/app.go`)
Lazy-initialized external clients:
- `Gotenberg`: PDF generation
- `Tika`: Document parsing/OCR
- `LLM`: Future analysis engine

Follows dependency injection pattern → easy to test with mocks.

## Development Conventions

| Aspect | Convention |
|--------|-----------|
| **Package names** | Lowercase, no underscores (`internal/formats/`, not `internal/Formats/`) |
| **Constructors** | `NewFoo(config Config) (*Foo, error)` |
| **Error wrapping** | `fmt.Errorf("context: %w", err)` — use `%w` verb |
| **Error codes** | Named constants: `ErrPDFImportDisabled`, `ErrCodeInvalidXML` |
| **Config precedence** | CLI flags > environment variables > hardcoded defaults |
| **IDs** | UUIDs throughout (google/uuid) for stable cross-referencing |
| **Timestamps** | `time.Time`, JSON marshals as RFC3339 string |
| **Validation** | Via embedded `schema.json` + `santhosh-tekuri/jsonschema` |
| **Element classification** | Type field is always a string constant from [model.ElementType](internal/model/types.go) |

## Common Pitfalls & Special Handling

### 1. PDF Import Fragility
- Poppler must be installed; check env var `SCREENJSON_PDFTOHTML`
- Returns `ErrPDFImportDisabled` if not available
- Rejects PDFs with <500 bytes extracted text (likely image/OCR-based, not searchable)
- External subprocess overhead — consider timeout handling in CLI

### 2. Character Management
- All character references use **UUID, not name** (first-class objects)
- Character name changes don't break references
- Bridge logic must populate both `doc.Characters` array AND create `charMap[UPPERCASE_NAME] = ID`
- Solves the "JOHN" vs "JOHN (CONT'D)" vs "JOHNNY" problem

### 3. Multi-language Text Fields
- Use `Text` type (map of language tag → string)
- Language tags: BCP 47 format, validated with regex `^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$`
- JSON marshals as object: `{"en": "...", "fr": "..."}`

### 4. Element Type Discriminator
- Every `Element` has a `Type` field (string)
- Type determines which other fields are populated:
  - `character` → `Character` field + `Display`
  - `dialogue` / `parenthetical` → `Text` field
  - `action` → `Text` field
  - Never mix types in one element

### 5. Format Auto-detection
- If `--format` not specified, registry uses file extension
- Extension matching is case-insensitive
- Falls back to explicit format specification if ambiguous

### 6. Async Job Processing
- REST API returns `202 Accepted` for async `/convert` and `/export`
- Job status progresses: `pending` → `running` → `completed` or `failed`
- Check job status via `GET /jobs/{id}`

## Key Exemplar Files

When implementing features, reference these files for patterns:

| File | Demonstrates |
|------|--------------|
| [cmd/screenjson/main.go](cmd/screenjson/main.go) | CLI dispatcher, all 7 commands |
| [internal/model/document.go](internal/model/document.go) | Complete domain model structure |
| [internal/model/elements.go](internal/model/elements.go) | Discriminated union pattern (Element type) |
| [internal/pipeline/registry.go](internal/pipeline/registry.go) | Format registry pattern with capabilities |
| [internal/formats/fountain/codec/decode.go](internal/formats/fountain/codec/decode.go) | Text format parsing with regex |
| [internal/formats/pdf/codec/decode.go](internal/formats/pdf/codec/decode.go) | External tool integration (Poppler) + geometry-based classification |
| [internal/formats/fountain/bridge/bridge.go](internal/formats/fountain/bridge/bridge.go) | Bridging format model to canonical Document |
| [internal/app/app.go](internal/app/app.go) | DI container with lazy initialization |
| [internal/server/server.go](internal/server/server.go) | REST API structure + error handling |

## Testing & Validation

**Schema validation**: All generated `Document` objects must pass validation against `internal/schema/schema.json`:
```go
validator := schema.NewValidator()
doc := &model.Document{ ... }
if err := validator.Validate(doc); err != nil {
    // non-nil means schema violation
}
```

**Example screenplay files** in `examples/`:
- `His_Girl_Friday_1940_Screenplay.fdx` (Final Draft)
- `His_Girl_Friday_1940_Screenplay.fountain` (Fountain)
- `His_Girl_Friday_1940_Screenplay.fadein` (FadeIn)
- `His_Girl_Friday_1940_Screenplay.fdx` (FDX)
- `His_Girl_Friday_1940_Screenplay.openscreenplay.document.xml` (OpenScreenplay)

## Links & Resources

- [screenjson.com](https://screenjson.com) — ScreenJSON format specification
- [screenjson-schema](../screenjson-schema/) — Canonical JSON Schema (Draft 2026-01)
- [Poppler Documentation](https://poppler.freedesktop.org/)
- [Fountain Specification](https://fountain.io/)

# screenjson-cli

A powerful command-line tool for converting, validating, and managing screenplay documents using the [ScreenJSON](https://screenjson.com) format.

## Features

- **Convert** screenplays from FDX (Final Draft), FadeIn, Fountain, and PDF to ScreenJSON
- **Export** ScreenJSON to FDX, FadeIn, Fountain, and PDF
- **Validate** documents against the ScreenJSON JSON Schema
- **Encrypt/Decrypt** content strings with AES-256 encryption
- **REST API** server for programmatic access
- **Multi-format output** (JSON and YAML)
- **Extensible storage** with database and blob store integrations

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Commands](#commands)
   - [convert](#convert)
   - [export](#export)
   - [validate](#validate)
   - [encrypt](#encrypt)
   - [decrypt](#decrypt)
   - [serve](#serve)
   - [formats](#formats)
   - [version](#version)
   - [help](#help)
4. [Supported Formats](#supported-formats)
5. [Environment Variables](#environment-variables)
6. [Docker](#docker)
7. [REST API](#rest-api)
8. [Examples](#examples)
9. [Troubleshooting](#troubleshooting)

---

## Installation

### From Source

```bash
git clone https://github.com/screenjson/screenjson-cli.git
cd screenjson-cli
go build -o screenjson ./cmd/screenjson/
sudo mv screenjson /usr/local/bin/
```

### Docker

```bash
docker build -t screenjson-cli .
docker run --rm screenjson-cli help
```

### PDF Import Requirement

PDF import requires [Poppler](https://poppler.freedesktop.org/) (specifically `pdftohtml`):

```bash
# macOS
brew install poppler

# Ubuntu/Debian
sudo apt-get install poppler-utils

# The binary is typically at:
# macOS (Intel): /usr/local/bin/pdftohtml
# macOS (Apple Silicon): /opt/homebrew/bin/pdftohtml
# Linux: /usr/bin/pdftohtml
```

---

## Quick Start

```bash
# Convert a Final Draft file to ScreenJSON
screenjson convert -i screenplay.fdx -o screenplay.json

# Convert Fountain to YAML
screenjson convert -i screenplay.fountain --yaml -o screenplay.yaml

# Export ScreenJSON to PDF
screenjson export -i screenplay.json -f pdf -o screenplay.pdf

# Validate a ScreenJSON document
screenjson validate -i screenplay.json

# Start the REST API server
screenjson serve --port 8080
```

---

## Commands

### convert

Convert screenplay formats to ScreenJSON.

```
screenjson convert -i <input-file> -o <output-file> [options]
```

#### Options

| Flag | Long Form | Description |
|------|-----------|-------------|
| `-i` | `--input` | Input file path **(required)** |
| `-o` | `--output` | Output file path (default: stdout) |
| `-f` | `--format` | Force input format: `fdx`, `fadein`, `fountain`, `celtx`, `pdf` |
| | `--yaml` | Output as YAML instead of JSON |
| | `--encrypt` | Encrypt content with shared secret (min 10 chars) |
| | `--lang` | Primary language code, BCP 47 (default: `en`) |
| `-q` | `--quiet` | Suppress non-error output |
| | `--verbose` | Enable verbose logging |

#### PDF Import Options

| Flag | Description |
|------|-------------|
| `--pdftohtml` | Path to `pdftohtml` binary |
| `--pdf-password` | PDF decryption password |

#### Database Output Options

| Flag | Description |
|------|-------------|
| `--db` | Database type: `elasticsearch`, `mongodb`, `cassandra`, `dynamodb`, `redis` |
| `--db-host` | Database host |
| `--db-port` | Database port |
| `--db-user` | Database username |
| `--db-pass` | Database password |
| `--db-collection` | Collection/index name |

#### Blob Storage Options

| Flag | Description |
|------|-------------|
| `--blob` | Blob store type: `s3`, `azure`, `minio` |
| `--blob-bucket` | Bucket name |
| `--blob-key` | Object key/path |
| `--blob-region` | AWS region (for S3) |
| `--blob-endpoint` | Custom endpoint (for Minio) |

#### Examples

```bash
# Basic conversion (auto-detect format from extension)
screenjson convert -i screenplay.fdx -o screenplay.json

# Force format detection
screenjson convert -i renamed_file.txt -f fountain -o screenplay.json

# Convert and encrypt
screenjson convert -i screenplay.fdx -o encrypted.json --encrypt "MySecretKey123"

# Convert to YAML
screenjson convert -i screenplay.fountain --yaml -o screenplay.yaml

# PDF import with custom pdftohtml path
screenjson convert -i screenplay.pdf --pdftohtml /usr/bin/pdftohtml -o screenplay.json

# Password-protected PDF
screenjson convert -i protected.pdf --pdf-password "secret" -o screenplay.json

# Output to stdout (for piping)
screenjson convert -i screenplay.fdx | jq '.title'

# Store directly to S3
screenjson convert -i screenplay.fdx --blob s3 --blob-bucket my-bucket --blob-region us-west-2
```

---

### export

Export ScreenJSON to screenplay formats.

```
screenjson export -i <input.json> -f <format> -o <output-file> [options]
```

#### Options

| Flag | Long Form | Description |
|------|-----------|-------------|
| `-i` | `--input` | Input ScreenJSON file **(required)** |
| `-o` | `--output` | Output file path (default: stdout) |
| `-f` | `--format` | Output format: `fdx`, `fadein`, `fountain`, `pdf` **(required)** |
| | `--decrypt` | Decrypt content before export |
| | `--lang` | Primary language for export (default: `en`) |
| `-q` | `--quiet` | Suppress non-error output |
| | `--verbose` | Enable verbose logging |

#### PDF Export Options

| Flag | Description |
|------|-------------|
| `--pdf-paper` | Paper size: `letter` (default), `a4` |
| `--pdf-font` | Font: `courier` (default), `courier-prime` |
| `--gotenberg` | Gotenberg URL for HTML→PDF rendering |

#### Examples

```bash
# Export to Final Draft
screenjson export -i screenplay.json -f fdx -o screenplay.fdx

# Export to Fountain
screenjson export -i screenplay.json -f fountain -o screenplay.fountain

# Export to FadeIn
screenjson export -i screenplay.json -f fadein -o screenplay.fadein

# Export to PDF
screenjson export -i screenplay.json -f pdf -o screenplay.pdf

# Export to PDF with A4 paper
screenjson export -i screenplay.json -f pdf -o screenplay.pdf --pdf-paper a4

# Export encrypted document (decrypt first)
screenjson export -i encrypted.json -f fdx -o screenplay.fdx --decrypt "MySecretKey123"

# Auto-detect format from output extension
screenjson export -i screenplay.json -o screenplay.fdx
```

---

### validate

Validate a ScreenJSON document against the schema.

```
screenjson validate -i <input.json> [options]
```

#### Options

| Flag | Long Form | Description |
|------|-----------|-------------|
| `-i` | `--input` | Input file path **(required)** |
| | `--strict` | Exit with error code on validation failure |
| | `--verbose` | Show schema version and additional info |

#### Examples

```bash
# Basic validation
screenjson validate -i screenplay.json

# Strict mode (for CI/CD)
screenjson validate -i screenplay.json --strict

# Verbose output
screenjson validate -i screenplay.json --verbose
```

#### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Valid document |
| 1 | Invalid document (with `--strict`) or error |

---

### encrypt

Encrypt content strings in a ScreenJSON document.

```
screenjson encrypt -i <input.json> -o <output.json> --key <secret>
```

#### Options

| Flag | Description |
|------|-------------|
| `-i`, `--input` | Input file path **(required)** |
| `-o`, `--output` | Output file path **(required)** |
| `--key` | Encryption key (min 10 characters) **(required)** |

The encryption uses **AES-256-CTR** with SHA-256 key derivation. Only text content is encrypted; structure, UUIDs, and metadata remain visible.

#### Examples

```bash
# Encrypt a document
screenjson encrypt -i screenplay.json -o encrypted.json --key "TopSecretKey2026!"

# Using environment variable for key
export SCREENJSON_ENCRYPT_KEY="TopSecretKey2026!"
screenjson encrypt -i screenplay.json -o encrypted.json
```

---

### decrypt

Decrypt content strings in an encrypted ScreenJSON document.

```
screenjson decrypt -i <input.json> -o <output.json> --key <secret>
```

#### Options

| Flag | Description |
|------|-------------|
| `-i`, `--input` | Input file path **(required)** |
| `-o`, `--output` | Output file path **(required)** |
| `--key` | Decryption key **(required)** |

#### Examples

```bash
# Decrypt a document
screenjson decrypt -i encrypted.json -o decrypted.json --key "TopSecretKey2026!"

# Using environment variable for key
export SCREENJSON_ENCRYPT_KEY="TopSecretKey2026!"
screenjson decrypt -i encrypted.json -o decrypted.json
```

---

### serve

Start the REST API server.

```
screenjson serve [options]
```

#### Options

| Flag | Description |
|------|-------------|
| `--host` | Listen host (default: `0.0.0.0`) |
| `--port` | Listen port (default: `8080`) |
| `--workers` | Worker pool size (default: CPU cores) |
| `--gotenberg` | Gotenberg URL for PDF rendering |
| `--tika` | Apache Tika URL for content extraction |
| `--llm` | LLM endpoint URL (OpenAI-compatible) |

#### Examples

```bash
# Start on default port
screenjson serve

# Custom port
screenjson serve --port 3000

# With external services
screenjson serve --gotenberg http://gotenberg:3000 --tika http://tika:9998

# Production configuration
screenjson serve --host 127.0.0.1 --port 8080 --workers 16
```

---

### formats

List all supported formats and their capabilities.

```
screenjson formats
```

#### Output

```
Supported Formats:
----------------------------------------------------------------------
fdx          Final Draft FDX format
             Extensions: .fdx
             Capabilities: import, export

fadein       FadeIn OSF format (ZIP with document.xml)
             Extensions: .fadein
             Capabilities: import, export

fountain     Fountain plain-text screenplay format
             Extensions: .fountain, .spmd
             Capabilities: import, export

pdf          PDF screenplay format
             Extensions: .pdf
             Capabilities: import, export

json         ScreenJSON native format
             Extensions: .json
             Capabilities: import, export

yaml         YAML serialization of ScreenJSON
             Extensions: .yaml, .yml
             Capabilities: import, export

celtx        Celtx project format (placeholder)
             Extensions: .celtx
             Capabilities: (not yet implemented)
```

---

### version

Display version information.

```
screenjson version
screenjson -v
screenjson --version
```

---

### help

Display help information.

```
screenjson help
screenjson -h
screenjson --help

# Command-specific help
screenjson convert --help
screenjson export --help
screenjson serve --help
```

---

## Supported Formats

| Format | Extension(s) | Import | Export | Notes |
|--------|--------------|--------|--------|-------|
| **Final Draft** | `.fdx` | ✅ | ✅ | XML-based |
| **FadeIn** | `.fadein` | ✅ | ✅ | ZIP archive with OSF XML |
| **Fountain** | `.fountain`, `.spmd` | ✅ | ✅ | Plain text with markup |
| **PDF** | `.pdf` | ✅ | ✅ | Import requires Poppler |
| **ScreenJSON** | `.json` | ✅ | ✅ | Native format |
| **YAML** | `.yaml`, `.yml` | ✅ | ✅ | ScreenJSON in YAML |
| **Celtx** | `.celtx` | ❌ | ❌ | Placeholder (coming soon) |

---

## Environment Variables

All configuration can be set via environment variables:

### General

| Variable | Description |
|----------|-------------|
| `SCREENJSON_ENCRYPT_KEY` | Default encryption/decryption key |
| `SCREENJSON_PDFTOHTML` | Path to `pdftohtml` binary |
| `SCREENJSON_PDF_PAPER` | Default paper size (`letter`, `a4`) |
| `SCREENJSON_PDF_FONT` | Default font (`courier`, `courier-prime`) |

### Database

| Variable | Description |
|----------|-------------|
| `SCREENJSON_DB_TYPE` | Database type |
| `SCREENJSON_DB_HOST` | Database host |
| `SCREENJSON_DB_PORT` | Database port |
| `SCREENJSON_DB_USER` | Database username |
| `SCREENJSON_DB_PASS` | Database password |
| `SCREENJSON_DB_COLLECTION` | Collection/index name |
| `SCREENJSON_DB_AUTH_TYPE` | Authentication type (`basic`, `apikey`, `token`) |
| `SCREENJSON_DB_APIKEY` | API key for authentication |
| `SCREENJSON_DB_INDEX` | Elasticsearch index name |
| `SCREENJSON_DB_REGION` | DynamoDB region |

### Blob Storage

| Variable | Description |
|----------|-------------|
| `SCREENJSON_BLOB_TYPE` | Blob store type (`s3`, `azure`, `minio`) |
| `SCREENJSON_BLOB_BUCKET` | Bucket name |
| `SCREENJSON_BLOB_KEY` | Object key/path |
| `SCREENJSON_BLOB_REGION` | AWS region |
| `SCREENJSON_BLOB_ENDPOINT` | Custom endpoint (Minio) |
| `SCREENJSON_AWS_ACCESS_KEY` | AWS access key ID |
| `SCREENJSON_AWS_SECRET_KEY` | AWS secret access key |

### External Services

| Variable | Description |
|----------|-------------|
| `SCREENJSON_GOTENBERG_URL` | Gotenberg URL |
| `SCREENJSON_TIKA_URL` | Apache Tika URL |
| `SCREENJSON_LLM_URL` | LLM endpoint URL |
| `SCREENJSON_LLM_APIKEY` | LLM API key |
| `SCREENJSON_LLM_MODEL` | LLM model name |

### Server

| Variable | Description |
|----------|-------------|
| `SCREENJSON_SERVER_HOST` | Server listen host |
| `SCREENJSON_SERVER_PORT` | Server listen port |
| `SCREENJSON_SERVER_WORKERS` | Worker pool size |

---

## Docker

### Build

```bash
docker build -t screenjson-cli .
```

### CLI Mode

```bash
# Help
docker run --rm screenjson-cli help

# Convert (mount local files)
docker run --rm -v $(pwd):/data screenjson-cli convert -i /data/screenplay.fdx -o /data/screenplay.json

# Validate
docker run --rm -v $(pwd):/data screenjson-cli validate -i /data/screenplay.json
```

### Server Mode

```bash
# Start server
docker run --rm -p 8080:8080 screenjson-cli serve --port 8080

# With environment configuration
docker run --rm -p 8080:8080 \
  -e SCREENJSON_GOTENBERG_URL=http://gotenberg:3000 \
  -e SCREENJSON_TIKA_URL=http://tika:9998 \
  screenjson-cli serve
```

### Docker Compose

```yaml
version: '3.8'
services:
  screenjson:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SCREENJSON_GOTENBERG_URL=http://gotenberg:3000
      - SCREENJSON_PDFTOHTML=/usr/bin/pdftohtml
    command: ["serve", "--port", "8080"]

  gotenberg:
    image: gotenberg/gotenberg:8
    ports:
      - "3000:3000"
```

---

## REST API

When running in server mode (`screenjson serve`), the following endpoints are available:

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Queue status and metrics |
| `GET` | `/health` | Health check |
| `GET` | `/formats` | List supported formats |
| `POST` | `/convert` | Convert file to ScreenJSON |
| `POST` | `/export` | Export ScreenJSON to format |
| `POST` | `/validate` | Validate ScreenJSON document |

### POST /convert

Convert a screenplay file to ScreenJSON.

**Request (multipart/form-data):**
```bash
curl -X POST http://localhost:8080/convert \
  -F "file=@screenplay.fdx" \
  -F "format=fdx" \
  -F "lang=en"
```

**Request (application/octet-stream):**
```bash
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/octet-stream" \
  -H "X-Input-Format: fdx" \
  --data-binary @screenplay.fdx
```

**Response:** ScreenJSON document (JSON)

### POST /export

Export ScreenJSON to a screenplay format.

**Request:**
```bash
curl -X POST http://localhost:8080/export \
  -H "Content-Type: application/json" \
  -H "X-Output-Format: fountain" \
  -d @screenplay.json
```

**Response:** Exported file (binary)

### POST /validate

Validate a ScreenJSON document.

**Request:**
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @screenplay.json
```

**Response:**
```json
{
  "valid": true,
  "errors": []
}
```

### GET /health

**Response:**
```json
{
  "status": "healthy",
  "version": "1.1.0",
  "uptime": "2h15m30s"
}
```

### GET /formats

**Response:**
```json
[
  {
    "name": "fdx",
    "extensions": [".fdx"],
    "mimeTypes": ["application/xml"],
    "canDecode": true,
    "canEncode": true,
    "description": "Final Draft FDX format"
  }
]
```

---

## Examples

### Complete Workflow: FDX to PDF

```bash
# 1. Convert Final Draft to ScreenJSON
screenjson convert -i screenplay.fdx -o screenplay.json

# 2. Validate the conversion
screenjson validate -i screenplay.json --strict

# 3. Export to PDF
screenjson export -i screenplay.json -f pdf -o screenplay.pdf
```

### Round-Trip Conversion

```bash
# Fountain → ScreenJSON → FDX → ScreenJSON
screenjson convert -i original.fountain -o step1.json
screenjson export -i step1.json -f fdx -o step2.fdx
screenjson convert -i step2.fdx -o step3.json

# Validate the round-trip
screenjson validate -i step3.json
```

### Batch Conversion

```bash
# Convert all .fdx files in a directory
for f in *.fdx; do
  screenjson convert -i "$f" -o "${f%.fdx}.json"
done
```

### Secure Workflow with Encryption

```bash
# Convert and encrypt in one step
screenjson convert -i screenplay.fdx -o encrypted.json --encrypt "MySecretKey123"

# Decrypt and export
screenjson export -i encrypted.json -f pdf -o screenplay.pdf --decrypt "MySecretKey123"
```

### CI/CD Validation

```bash
#!/bin/bash
# validate-screenplays.sh

set -e

for file in screenplays/*.json; do
  echo "Validating $file..."
  screenjson validate -i "$file" --strict
done

echo "All screenplays valid!"
```

### Piping and Streaming

```bash
# Extract title from screenplay
screenjson convert -i screenplay.fdx | jq -r '.title.en'

# Convert and upload to S3 in one pipeline
screenjson convert -i screenplay.fdx | aws s3 cp - s3://my-bucket/screenplay.json

# Chain conversions
cat screenplay.fountain | screenjson convert -i /dev/stdin -f fountain | screenjson export -f fdx -o screenplay.fdx
```

---

## Troubleshooting

### PDF Import Not Working

```
Error: PDF import requires pdftohtml (Poppler)
```

**Solution:** Install Poppler and specify the path:
```bash
# macOS
brew install poppler
screenjson convert -i file.pdf --pdftohtml /opt/homebrew/bin/pdftohtml -o file.json

# Or set environment variable
export SCREENJSON_PDFTOHTML=/opt/homebrew/bin/pdftohtml
```

### File Not Found

```
Error: file does not exist: screenplay.fdx
```

**Solution:** Use absolute paths or check your current directory:
```bash
screenjson convert -i /full/path/to/screenplay.fdx -o output.json
```

### Permission Denied

```
Error: file is not readable: screenplay.fdx
```

**Solution:** Check file permissions:
```bash
chmod 644 screenplay.fdx
```

### Invalid JSON/YAML Input

```
Error parsing input (not valid JSON or YAML)
```

**Solution:** Ensure the input file is valid ScreenJSON:
```bash
# Check JSON syntax
jq . screenplay.json

# Validate structure
screenjson validate -i screenplay.json
```

### Encryption Key Too Short

```
Error: encryption key must be at least 10 characters
```

**Solution:** Use a longer key:
```bash
screenjson encrypt -i file.json -o encrypted.json --key "AtLeast10Chars!"
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Links

- [ScreenJSON Specification](https://screenjson.com)
- [JSON Schema](https://screenjson.com/draft/2026-04/schema)
- [Issue Tracker](https://github.com/screenjson/screenjson-cli/issues)

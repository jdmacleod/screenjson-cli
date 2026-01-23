// Package main provides the ScreenJSON CLI application.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"screenjson/cli/internal/app"
	"screenjson/cli/internal/config"
	"screenjson/cli/internal/crypto"
	fdxbridge "screenjson/cli/internal/formats/fdx/bridge"
	fdxcodec "screenjson/cli/internal/formats/fdx/codec"
	fadeinbridge "screenjson/cli/internal/formats/fadein/bridge"
	fadeincodec "screenjson/cli/internal/formats/fadein/codec"
	fountainbridge "screenjson/cli/internal/formats/fountain/bridge"
	fountaincodec "screenjson/cli/internal/formats/fountain/codec"
	pdfcodec "screenjson/cli/internal/formats/pdf/codec"
	"screenjson/cli/internal/model"
	"screenjson/cli/internal/pipeline"
	"screenjson/cli/internal/schema"
	"screenjson/cli/internal/server"
)

var (
	version = "1.0.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "convert":
		cmdConvert(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "validate":
		cmdValidate(os.Args[2:])
	case "encrypt":
		cmdEncrypt(os.Args[2:])
	case "decrypt":
		cmdDecrypt(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "formats":
		cmdFormats()
	case "version", "-v", "--version":
		fmt.Println("screenjson version", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`screenjson - Screenplay format converter and ScreenJSON toolkit

USAGE:
    screenjson <command> [options]

COMMANDS:
    convert     Convert screenplay format to ScreenJSON
    export      Convert ScreenJSON to screenplay format
    validate    Validate a ScreenJSON file against schema
    encrypt     Encrypt content strings in a ScreenJSON file
    decrypt     Decrypt content strings in a ScreenJSON file
    serve       Start REST API server
    formats     List supported formats
    version     Show version information
    help        Show this help

GLOBAL OPTIONS:
    --help, -h              Show help
    --version, -v           Show version
    --quiet, -q             Suppress non-error output
    --verbose               Enable verbose logging

Use "screenjson <command> --help" for command-specific options.

ENVIRONMENT VARIABLES:
    SCREENJSON_DB_TYPE          Database type
    SCREENJSON_DB_HOST          Database host
    SCREENJSON_DB_PORT          Database port
    SCREENJSON_DB_USER          Database username
    SCREENJSON_DB_PASS          Database password
    SCREENJSON_DB_COLLECTION    Collection name
    SCREENJSON_BLOB_TYPE        Blob store type
    SCREENJSON_BLOB_BUCKET      Blob bucket
    SCREENJSON_AWS_ACCESS_KEY   AWS access key
    SCREENJSON_AWS_SECRET_KEY   AWS secret key
    SCREENJSON_AWS_REGION       AWS region
    SCREENJSON_GOTENBERG_URL    Gotenberg URL
    SCREENJSON_TIKA_URL         Apache Tika URL
    SCREENJSON_LLM_URL          LLM endpoint URL
    SCREENJSON_ENCRYPT_KEY      Default encryption key
    SCREENJSON_PDFTOHTML        Path to pdftohtml

EXAMPLES:
    screenjson convert -i screenplay.fdx -o screenplay.json
    screenjson convert -i screenplay.fountain --yaml -o screenplay.yaml
    screenjson export -i screenplay.json -f fdx -o screenplay.fdx
    screenjson export -i screenplay.json -f pdf -o screenplay.pdf
    screenjson validate -i screenplay.json
    screenjson encrypt -i screenplay.json -o encrypted.json --key "mypassword"
    screenjson serve --port 8080`)
}

// cmdConvert handles the convert command (format -> ScreenJSON).
func cmdConvert(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	input := fs.String("i", "", "Input file path (required)")
	inputLong := fs.String("input", "", "Input file path (required)")
	output := fs.String("o", "", "Output file (default: stdout)")
	outputLong := fs.String("output", "", "Output file (default: stdout)")
	fromFormat := fs.String("f", "", "Force input format (fdx|fadein|fountain|celtx|pdf)")
	fromFormatLong := fs.String("format", "", "Force input format (fdx|fadein|fountain|celtx|pdf)")
	useYAML := fs.Bool("yaml", false, "Output as YAML instead of JSON")
	encryptKey := fs.String("encrypt", "", "Encrypt content with shared secret (min 10 chars)")
	lang := fs.String("lang", "en", "Primary language code (BCP 47)")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	quiet := fs.Bool("q", false, "Suppress non-error output")
	quietLong := fs.Bool("quiet", false, "Suppress non-error output")

	// Database output options
	dbType := fs.String("db", "", "Database type (elasticsearch|mongodb|cassandra|dynamodb|redis)")
	dbHost := fs.String("db-host", "", "Database host")
	dbPort := fs.Int("db-port", 0, "Database port")
	dbUser := fs.String("db-user", "", "Database username")
	dbPass := fs.String("db-pass", "", "Database password")
	dbCollection := fs.String("db-collection", "", "Collection/index name")

	// Blob output options
	blobType := fs.String("blob", "", "Blob store (s3|azure|minio)")
	blobBucket := fs.String("blob-bucket", "", "Bucket name")
	blobKey := fs.String("blob-key", "", "Object key/path")
	blobRegion := fs.String("blob-region", "", "AWS region (S3)")
	blobEndpoint := fs.String("blob-endpoint", "", "Custom endpoint (Minio)")

	// PDF import options
	pdfToHTML := fs.String("pdftohtml", "", "Path to pdftohtml binary")
	pdfPassword := fs.String("pdf-password", "", "PDF decryption password")

	fs.Usage = func() {
		fmt.Println(`CONVERT (format -> ScreenJSON):
    screenjson convert -i <input-file> -o <output-file> [options]

    Converts screenplay formats (FDX, FadeIn, Fountain, PDF) to ScreenJSON.

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
    screenjson convert -i screenplay.fdx -o screenplay.json
    screenjson convert -i screenplay.fountain --yaml -o screenplay.yaml
    screenjson convert -i screenplay.pdf --pdftohtml /opt/homebrew/bin/pdftohtml -o screenplay.json`)
	}

	fs.Parse(args)

	inputFile := coalesce(*input, *inputLong)
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file required (-i/--input)")
		fs.Usage()
		os.Exit(1)
	}

	outputFile := coalesce(*output, *outputLong)
	format := coalesce(*fromFormat, *fromFormatLong)
	isQuiet := *quiet || *quietLong

	// Auto-detect format
	if format == "" {
		if info, ok := pipeline.GlobalRegistry.GetByFilename(inputFile); ok {
			format = info.Name
		} else {
			fmt.Fprintln(os.Stderr, "Error: could not detect input format, use -f/--format")
			os.Exit(1)
		}
	}
	if info, ok := pipeline.GlobalRegistry.Get(format); !ok || !info.CanDecode() {
		fmt.Fprintf(os.Stderr, "Error: unsupported input format: %s\n", format)
		os.Exit(1)
	}

	if *verbose && !isQuiet {
		fmt.Fprintf(os.Stderr, "Converting %s (%s) -> ScreenJSON\n", inputFile, format)
	}

	// Read input
	inputData, err := requireReadableFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Convert to ScreenJSON
	doc, err := convertToScreenJSON(inputData, format, *lang, *pdfToHTML, *pdfPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}

	// Encrypt if requested
	if *encryptKey != "" {
		if err := crypto.EncryptDocument(doc, *encryptKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error encrypting: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate
	validator, err := schema.NewValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create validator: %v\n", err)
	} else {
		jsonData, _ := json.Marshal(doc)
		if result, err := validator.Validate(jsonData); err == nil && !result.Valid {
			fmt.Fprintln(os.Stderr, "Warning: converted document has validation errors")
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Path, e.Message)
			}
		}
	}

	// Encode output
	var outputData []byte
	if *useYAML {
		outputData, err = yaml.Marshal(doc)
	} else {
		outputData, err = json.MarshalIndent(doc, "", "  ")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}

	// Handle database output
	if *dbType != "" {
		if err := storeToDatabase(doc, *dbType, *dbHost, *dbPort, *dbUser, *dbPass, *dbCollection); err != nil {
			fmt.Fprintf(os.Stderr, "Error storing to database: %v\n", err)
			os.Exit(1)
		}
		if !isQuiet {
			fmt.Fprintf(os.Stderr, "Stored document %s to %s\n", doc.ID, *dbType)
		}
	}

	// Handle blob output
	if *blobType != "" {
		key := *blobKey
		if key == "" {
			key = doc.ID + ".json"
			if *useYAML {
				key = doc.ID + ".yaml"
			}
		}
		contentType := "application/json"
		if *useYAML {
			contentType = "application/x-yaml"
		}
		if err := storeToBlob(outputData, *blobType, *blobBucket, key, *blobRegion, *blobEndpoint, contentType); err != nil {
			fmt.Fprintf(os.Stderr, "Error storing to blob: %v\n", err)
			os.Exit(1)
		}
		if !isQuiet {
			fmt.Fprintf(os.Stderr, "Stored document to %s://%s/%s\n", *blobType, *blobBucket, key)
		}
	}

	// Write to file or stdout
	if outputFile == "" || outputFile == "-" {
		fmt.Print(string(outputData))
	} else {
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		if *verbose && !isQuiet {
			fmt.Fprintf(os.Stderr, "Wrote %d bytes to %s\n", len(outputData), outputFile)
		}
	}
}

// cmdExport handles the export command (ScreenJSON -> format).
func cmdExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	input := fs.String("i", "", "Input file path (required)")
	inputLong := fs.String("input", "", "Input file path (required)")
	output := fs.String("o", "", "Output file (default: stdout)")
	outputLong := fs.String("output", "", "Output file (default: stdout)")
	toFormat := fs.String("f", "", "Output format (fdx|fadein|fountain|pdf) [required]")
	toFormatLong := fs.String("format", "", "Output format (fdx|fadein|fountain|pdf) [required]")
	decryptKey := fs.String("decrypt", "", "Decrypt content before export")
	lang := fs.String("lang", "en", "Primary language for export")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	quiet := fs.Bool("q", false, "Suppress non-error output")
	quietLong := fs.Bool("quiet", false, "Suppress non-error output")

	// PDF export options
	pdfPaper := fs.String("pdf-paper", "letter", "Paper size (letter|a4)")
	pdfFont := fs.String("pdf-font", "courier", "Font (courier|courier-prime)")

	// External services
	gotenbergURL := fs.String("gotenberg", "", "Gotenberg URL for HTML->PDF rendering")

	fs.Usage = func() {
		fmt.Println(`EXPORT (ScreenJSON -> format):
    screenjson export -i <input.json> -f <format> -o <output-file> [options]

    Exports ScreenJSON to screenplay formats (FDX, FadeIn, Fountain, PDF).

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
    screenjson export -i screenplay.json -f fdx -o screenplay.fdx
    screenjson export -i screenplay.json -f fountain -o screenplay.fountain
    screenjson export -i screenplay.json -f pdf -o screenplay.pdf --pdf-paper letter`)
	}

	fs.Parse(args)

	inputFile := coalesce(*input, *inputLong)
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file required (-i/--input)")
		fs.Usage()
		os.Exit(1)
	}

	outputFile := coalesce(*output, *outputLong)
	format := coalesce(*toFormat, *toFormatLong)
	isQuiet := *quiet || *quietLong

	if format == "" {
		// Try to detect from output filename
		if outputFile != "" && outputFile != "-" {
			if info, ok := pipeline.GlobalRegistry.GetByFilename(outputFile); ok {
				format = info.Name
			}
		}
		if format == "" {
			fmt.Fprintln(os.Stderr, "Error: output format required (-f/--format)")
			fs.Usage()
			os.Exit(1)
		}
	}
	if info, ok := pipeline.GlobalRegistry.Get(format); !ok || !info.CanEncode() {
		fmt.Fprintf(os.Stderr, "Error: unsupported output format: %s\n", format)
		os.Exit(1)
	}

	if *verbose && !isQuiet {
		fmt.Fprintf(os.Stderr, "Exporting %s -> %s\n", inputFile, format)
	}

	// Read and parse input
	inputData, err := requireReadableFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var doc model.Document
	
	// Try JSON first, then YAML
	if err := json.Unmarshal(inputData, &doc); err != nil {
		if err := yaml.Unmarshal(inputData, &doc); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input (not valid JSON or YAML): %v\n", err)
			os.Exit(1)
		}
	}

	// Decrypt if requested
	if *decryptKey != "" {
		if err := crypto.DecryptDocument(&doc, *decryptKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error decrypting: %v\n", err)
			os.Exit(1)
		}
	}

	// Export to format
	outputData, err := exportFromScreenJSON(&doc, format, *lang, *pdfPaper, *pdfFont, *gotenbergURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if outputFile == "" || outputFile == "-" {
		os.Stdout.Write(outputData)
	} else {
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
		if *verbose && !isQuiet {
			fmt.Fprintf(os.Stderr, "Wrote %d bytes to %s\n", len(outputData), outputFile)
		}
	}

	// Suppress unused warnings for now
	_ = gotenbergURL
}

// cmdValidate handles the validate command.
func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	input := fs.String("i", "", "Input file path (required)")
	inputLong := fs.String("input", "", "Input file path (required)")
	strict := fs.Bool("strict", false, "Fail on warnings")
	verbose := fs.Bool("verbose", false, "Verbose output")

	fs.Usage = func() {
		fmt.Println(`VALIDATE:
    screenjson validate -i <input.json>

    Validates a ScreenJSON document against the schema.

Options:`)
		fs.PrintDefaults()
	}

	fs.Parse(args)

	inputFile := coalesce(*input, *inputLong)
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file required (-i/--input)")
		fs.Usage()
		os.Exit(1)
	}

	// Read input
	data, err := requireReadableFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Validate
	validator, err := schema.NewValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
		os.Exit(1)
	}

	result, err := validator.Validate(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Println("✓ Document is valid")
		if *verbose {
			fmt.Printf("Schema version: %s\n", schema.SchemaVersion)
		}
	} else {
		fmt.Println("✗ Document is invalid")
		for _, e := range result.Errors {
			fmt.Printf("  - %s: %s\n", e.Path, e.Message)
		}
		if *strict {
			os.Exit(1)
		}
	}
	
	_ = strict // Suppress unused warning
}

// cmdEncrypt handles the encrypt command.
func cmdEncrypt(args []string) {
	fs := flag.NewFlagSet("encrypt", flag.ExitOnError)
	input := fs.String("i", "", "Input file path (required)")
	inputLong := fs.String("input", "", "Input file path (required)")
	output := fs.String("o", "", "Output file (required)")
	outputLong := fs.String("output", "", "Output file (required)")
	key := fs.String("key", "", "Encryption key (min 10 chars, required)")

	fs.Usage = func() {
		fmt.Println(`ENCRYPT:
    screenjson encrypt -i <input.json> -o <output.json> --key <secret>

    Encrypts content strings in a ScreenJSON document using AES-256.

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
The encryption uses AES-256-CTR with SHA-256 key derivation.
Only text content is encrypted; structure and UUIDs remain visible.`)
	}

	fs.Parse(args)

	inputFile := coalesce(*input, *inputLong)
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file required (-i/--input)")
		fs.Usage()
		os.Exit(1)
	}

	outputFile := coalesce(*output, *outputLong)

	if outputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: output file required (-o/--output)")
		fs.Usage()
		os.Exit(1)
	}

	if *key == "" {
		*key = os.Getenv("SCREENJSON_ENCRYPT_KEY")
	}
	if *key == "" {
		fmt.Fprintln(os.Stderr, "Error: encryption key required (--key or SCREENJSON_ENCRYPT_KEY)")
		os.Exit(1)
	}

	// Read and parse input
	data, err := requireReadableFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var doc model.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Encrypt
	if err := crypto.EncryptDocument(&doc, *key); err != nil {
		fmt.Fprintf(os.Stderr, "Error encrypting: %v\n", err)
		os.Exit(1)
	}

	// Write output
	outputData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Document encrypted successfully")
}

// cmdDecrypt handles the decrypt command.
func cmdDecrypt(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	input := fs.String("i", "", "Input file path (required)")
	inputLong := fs.String("input", "", "Input file path (required)")
	output := fs.String("o", "", "Output file (required)")
	outputLong := fs.String("output", "", "Output file (required)")
	key := fs.String("key", "", "Decryption key (required)")

	fs.Usage = func() {
		fmt.Println(`DECRYPT:
    screenjson decrypt -i <input.json> -o <output.json> --key <secret>

    Decrypts content strings in a ScreenJSON document.

Options:`)
		fs.PrintDefaults()
	}

	fs.Parse(args)

	inputFile := coalesce(*input, *inputLong)
	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: input file required (-i/--input)")
		fs.Usage()
		os.Exit(1)
	}

	outputFile := coalesce(*output, *outputLong)

	if outputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: output file required (-o/--output)")
		fs.Usage()
		os.Exit(1)
	}

	if *key == "" {
		*key = os.Getenv("SCREENJSON_ENCRYPT_KEY")
	}
	if *key == "" {
		fmt.Fprintln(os.Stderr, "Error: decryption key required (--key or SCREENJSON_ENCRYPT_KEY)")
		os.Exit(1)
	}

	// Read and parse input
	data, err := requireReadableFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var doc model.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Decrypt
	if err := crypto.DecryptDocument(&doc, *key); err != nil {
		fmt.Fprintf(os.Stderr, "Error decrypting: %v\n", err)
		os.Exit(1)
	}

	// Write output
	outputData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Document decrypted successfully")
}

// cmdServe handles the serve command.
func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	host := fs.String("host", "0.0.0.0", "Listen host")
	port := fs.Int("port", 8080, "Listen port")
	workers := fs.Int("workers", 0, "Worker pool size (default: CPU cores)")

	// External services
	gotenbergURL := fs.String("gotenberg", "", "Gotenberg URL")
	tikaURL := fs.String("tika", "", "Apache Tika URL")
	llmURL := fs.String("llm", "", "LLM endpoint URL (OpenAI-compatible)")

	fs.Usage = func() {
		fmt.Println(`SERVE (REST API):
    screenjson serve [options]

    Starts the REST API server.

Options:`)
		fs.PrintDefaults()
		fmt.Println(`
Endpoints:
    GET  /               Queue status and metrics
    POST /convert        Upload file -> ScreenJSON response
    POST /export         Upload ScreenJSON -> format response
    GET  /health         Health check
    GET  /formats        List supported formats
    POST /validate       Validate ScreenJSON document`)
	}

	fs.Parse(args)

	// Create config
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()
	cfg.Server.Host = *host
	cfg.Server.Port = *port
	cfg.Server.Workers = *workers

	if *gotenbergURL != "" {
		cfg.Gotenberg.URL = *gotenbergURL
	}
	if *tikaURL != "" {
		cfg.Tika.URL = *tikaURL
	}
	if *llmURL != "" {
		cfg.LLM.URL = *llmURL
	}

	// Create app
	application, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	// Start background services
	application.Start()
	defer application.Stop()

	// Create and start server
	srv := server.New(application)

	// Handle shutdown
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		fmt.Println("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		}
		close(done)
	}()

	fmt.Printf("Server starting on http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}

	<-done
	fmt.Println("Server stopped")
}

// cmdFormats handles the formats command.
func cmdFormats() {
	formats := pipeline.GlobalRegistry.List()

	fmt.Println("Supported Formats:")
	fmt.Println(strings.Repeat("-", 70))

	for _, f := range formats {
		caps := []string{}
		if f.CanDecode() {
			caps = append(caps, "import")
		}
		if f.CanEncode() {
			caps = append(caps, "export")
		}

		fmt.Printf("%-12s %s\n", f.Name, f.Description)
		fmt.Printf("             Extensions: %s\n", strings.Join(f.Extensions, ", "))
		fmt.Printf("             Capabilities: %s\n", strings.Join(caps, ", "))
		fmt.Println()
	}
}

// convertToScreenJSON converts input data to a ScreenJSON document.
func convertToScreenJSON(data []byte, format, lang, pdfToHTML, pdfPassword string) (*model.Document, error) {
	ctx := context.Background()

	switch format {
	case "fdx":
		decoder := fdxcodec.NewDecoder()
		fdx, err := decoder.Decode(ctx, data)
		if err != nil {
			return nil, err
		}
		return fdxbridge.ToScreenJSON(fdx, lang), nil

	case "fadein":
		decoder := fadeincodec.NewDecoder()
		osf, err := decoder.Decode(ctx, data)
		if err != nil {
			return nil, err
		}
		return fadeinbridge.ToScreenJSON(osf, lang), nil

	case "fountain":
		decoder := fountaincodec.NewDecoder()
		ftn, err := decoder.Decode(ctx, data)
		if err != nil {
			return nil, err
		}
		return fountainbridge.ToScreenJSON(ftn, lang), nil

	case "pdf":
		if pdfToHTML == "" {
			pdfToHTML = os.Getenv("SCREENJSON_PDFTOHTML")
		}
		if pdfToHTML == "" {
			pdfToHTML = "/opt/homebrew/bin/pdftohtml"
		}
		decoder := pdfcodec.NewDecoder(pdfToHTML)
		if !decoder.IsAvailable() {
			return nil, fmt.Errorf("PDF import requires pdftohtml (Poppler). Install Poppler or specify --pdftohtml path")
		}
		return decoder.Decode(ctx, data, pdfPassword)

	case "json":
		var doc model.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		return &doc, nil

	case "yaml":
		var doc model.Document
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		return &doc, nil

	case "celtx":
		return nil, fmt.Errorf("Celtx import not yet implemented (placeholder)")

	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// exportFromScreenJSON exports a ScreenJSON document to a format.
func exportFromScreenJSON(doc *model.Document, format, lang, pdfPaper, pdfFont, gotenbergURL string) ([]byte, error) {
	ctx := context.Background()

	switch format {
	case "fdx":
		fdx := fdxbridge.FromScreenJSON(doc, lang)
		encoder := fdxcodec.NewEncoder()
		return encoder.Encode(ctx, fdx)

	case "fadein":
		osf := fadeinbridge.FromScreenJSON(doc, lang)
		encoder := fadeincodec.NewEncoder()
		return encoder.Encode(ctx, osf)

	case "fountain":
		ftn := fountainbridge.FromScreenJSON(doc, lang)
		encoder := fountaincodec.NewEncoder()
		return encoder.Encode(ctx, ftn)

	case "pdf":
		encoder := pdfcodec.NewEncoder()
		return encoder.Encode(ctx, doc)

	case "json":
		return json.MarshalIndent(doc, "", "  ")

	case "yaml":
		return yaml.Marshal(doc)

	case "celtx":
		return nil, fmt.Errorf("Celtx export not yet implemented (placeholder)")

	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// storeToDatabase stores a document to a database.
func storeToDatabase(doc *model.Document, dbType, host string, port int, user, pass, collection string) error {
	// This is a placeholder - actual implementation would use the stores package
	_ = doc
	_ = host
	_ = port
	_ = user
	_ = pass
	_ = collection
	return fmt.Errorf("database storage not yet wired up for %s", dbType)
}

// storeToBlob stores data to blob storage.
func storeToBlob(data []byte, blobType, bucket, key, region, endpoint, contentType string) error {
	// This is a placeholder - actual implementation would use the blob package
	_ = data
	_ = bucket
	_ = key
	_ = region
	_ = endpoint
	_ = contentType
	return fmt.Errorf("blob storage not yet wired up for %s", blobType)
}

// coalesce returns the first non-empty string.
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// requireReadableFile returns file contents, or a clear error if unreadable.
func requireReadableFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("input path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file does not exist: %s", path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, fmt.Errorf("file is not readable: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return os.ReadFile(path)
}

// getFilenameWithoutExt returns filename without extension.
func getFilenameWithoutExt(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

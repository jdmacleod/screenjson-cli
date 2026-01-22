// Package server provides a REST API server for ScreenJSON operations.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"screenjson/cli/internal/app"
	fdxbridge "screenjson/cli/internal/formats/fdx/bridge"
	fdxcodec "screenjson/cli/internal/formats/fdx/codec"
	fadeinbridge "screenjson/cli/internal/formats/fadein/bridge"
	fadeincodec "screenjson/cli/internal/formats/fadein/codec"
	fountainbridge "screenjson/cli/internal/formats/fountain/bridge"
	fountaincodec "screenjson/cli/internal/formats/fountain/codec"
	pdfcodec "screenjson/cli/internal/formats/pdf/codec"
	"screenjson/cli/internal/model"
	"screenjson/cli/internal/pipeline"
)

// Server is the REST API server.
type Server struct {
	app    *app.App
	server *http.Server
}

// New creates a new REST API server.
func New(application *app.App) *Server {
	s := &Server{
		app: application,
	}

	mux := http.NewServeMux()

	// Status and health
	mux.HandleFunc("GET /", s.handleRoot)
	mux.HandleFunc("GET /health", s.handleHealth)

	// Format info
	mux.HandleFunc("GET /formats", s.handleFormats)

	// Conversion endpoints
	mux.HandleFunc("POST /convert", s.handleConvert)
	mux.HandleFunc("POST /export", s.handleExport)
	mux.HandleFunc("POST /validate", s.handleValidate)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", application.Config.Server.Host, application.Config.Server.Port),
		Handler:      corsMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// Start starts the server.
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleRoot serves queue status and metrics.
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "ScreenJSON CLI API",
		"version": "1.0.0",
		"status":  "running",
		"queue": map[string]interface{}{
			"pending": s.app.Queue.Pending(),
			"workers": s.app.Queue.Workers(),
		},
		"endpoints": []map[string]string{
			{"method": "GET", "path": "/", "description": "Queue status and metrics"},
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/formats", "description": "List supported formats"},
			{"method": "POST", "path": "/convert", "description": "Upload file -> ScreenJSON response"},
			{"method": "POST", "path": "/export", "description": "Upload ScreenJSON -> format response"},
			{"method": "POST", "path": "/validate", "description": "Validate ScreenJSON document"},
		},
	}
	writeJSON(w, http.StatusOK, response)
}

// handleHealth returns service health status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "ok",
		"queue": map[string]interface{}{
			"pending": s.app.Queue.Pending(),
			"workers": s.app.Queue.Workers(),
		},
	}

	// Check external services
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := s.app.HealthCheck(ctx)
	servicesStatus := make(map[string]string)
	for name, err := range services {
		if err != nil {
			servicesStatus[name] = "error: " + err.Error()
		} else {
			servicesStatus[name] = "ok"
		}
	}
	if len(servicesStatus) > 0 {
		health["services"] = servicesStatus
	}

	writeJSON(w, http.StatusOK, health)
}

// handleFormats returns supported formats.
func (s *Server) handleFormats(w http.ResponseWriter, r *http.Request) {
	formats := pipeline.GlobalRegistry.List()

	response := make([]map[string]interface{}, 0, len(formats))
	for _, f := range formats {
		response = append(response, map[string]interface{}{
			"name":        f.Name,
			"extensions":  f.Extensions,
			"mimeTypes":   f.MIMETypes,
			"canDecode":   f.CanDecode(),
			"canEncode":   f.CanEncode(),
			"description": f.Description,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// handleConvert handles file upload and conversion to ScreenJSON.
func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB limit

	// Check content type
	contentType := r.Header.Get("Content-Type")

	var inputData []byte
	var fromFormat, lang string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Multipart form upload
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "No file provided")
			return
		}
		defer file.Close()

		inputData, err = io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to read file: "+err.Error())
			return
		}

		fromFormat = r.FormValue("format")
		lang = r.FormValue("lang")

		// Auto-detect format from filename if not specified
		if fromFormat == "" {
			if info, ok := pipeline.GlobalRegistry.GetByFilename(header.Filename); ok {
				fromFormat = info.Name
			}
		}
	} else {
		// Raw body with format in query params
		var err error
		inputData, err = io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to read body: "+err.Error())
			return
		}
		fromFormat = r.URL.Query().Get("format")
		lang = r.URL.Query().Get("lang")
	}

	if fromFormat == "" {
		writeError(w, http.StatusBadRequest, "Format not specified (use ?format= or form field)")
		return
	}
	if _, ok := pipeline.GlobalRegistry.Get(fromFormat); !ok {
		writeError(w, http.StatusBadRequest, "Unsupported input format: "+fromFormat)
		return
	}
	if lang == "" {
		lang = "en"
	}
	if len(inputData) == 0 {
		writeError(w, http.StatusBadRequest, "Empty request body")
		return
	}

	// Convert to ScreenJSON
	doc, err := convertToScreenJSON(ctx, inputData, fromFormat, lang, s.app.Config.PDF.PdfToHtml, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Conversion error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

// handleExport handles ScreenJSON to format export.
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50MB limit

	// Read input
	inputData, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read body: "+err.Error())
		return
	}

	// Parse ScreenJSON
	var doc model.Document
	if err := json.Unmarshal(inputData, &doc); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(inputData, &doc); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON/YAML: "+err.Error())
			return
		}
	}

	// Get format from query params
	toFormat := r.URL.Query().Get("format")
	lang := r.URL.Query().Get("lang")
	if toFormat == "" {
		writeError(w, http.StatusBadRequest, "Format not specified (use ?format=)")
		return
	}
	if _, ok := pipeline.GlobalRegistry.Get(toFormat); !ok {
		writeError(w, http.StatusBadRequest, "Unsupported output format: "+toFormat)
		return
	}
	if lang == "" {
		lang = "en"
	}

	// Export to format
	outputData, err := exportFromScreenJSON(ctx, &doc, toFormat, lang)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Export error: "+err.Error())
		return
	}

	// Set content type based on format
	contentType := pipeline.ContentTypeForFormat(toFormat)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s%s\"", doc.ID, pipeline.ExtensionForFormat(toFormat)))
	w.WriteHeader(http.StatusOK)
	w.Write(outputData)
}

// handleValidate validates a ScreenJSON document.
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB limit
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read body")
		return
	}
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, "Empty request body")
		return
	}

	result, err := s.app.Validator.Validate(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"valid": result.Valid,
	}
	if !result.Valid {
		errors := make([]map[string]string, 0, len(result.Errors))
		for _, e := range result.Errors {
			errors = append(errors, map[string]string{
				"path":    e.Path,
				"message": e.Message,
			})
		}
		response["errors"] = errors
	}

	writeJSON(w, http.StatusOK, response)
}

// convertToScreenJSON converts input data to a ScreenJSON document.
func convertToScreenJSON(ctx context.Context, data []byte, format, lang, pdfToHTML, pdfPassword string) (*model.Document, error) {
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
		decoder := pdfcodec.NewDecoder(pdfToHTML)
		if !decoder.IsAvailable() {
			return nil, fmt.Errorf("PDF import requires pdftohtml (Poppler)")
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

	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// exportFromScreenJSON exports a ScreenJSON document to a format.
func exportFromScreenJSON(ctx context.Context, doc *model.Document, format, lang string) ([]byte, error) {
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

	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// loggingMiddleware logs requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// corsMiddleware adds CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

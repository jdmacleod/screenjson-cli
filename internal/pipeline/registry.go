// Package pipeline provides format registration and document processing.
package pipeline

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Capability defines what a format driver can do.
type Capability int

const (
	CapDecode Capability = 1 << iota // Can decode (import)
	CapEncode                        // Can encode (export)
)

// FormatInfo describes a registered format.
type FormatInfo struct {
	Name         string
	Extensions   []string
	MIMETypes    []string
	Capabilities Capability
	Description  string
}

// CanDecode returns true if the format supports decoding.
func (f *FormatInfo) CanDecode() bool {
	return f.Capabilities&CapDecode != 0
}

// CanEncode returns true if the format supports encoding.
func (f *FormatInfo) CanEncode() bool {
	return f.Capabilities&CapEncode != 0
}

// Registry holds all registered format drivers.
type Registry struct {
	mu      sync.RWMutex
	formats map[string]*FormatInfo
	byExt   map[string]*FormatInfo
	byMIME  map[string]*FormatInfo
}

// GlobalRegistry is the global format registry.
var GlobalRegistry = NewRegistry()

// NewRegistry creates a new format registry.
func NewRegistry() *Registry {
	return &Registry{
		formats: make(map[string]*FormatInfo),
		byExt:   make(map[string]*FormatInfo),
		byMIME:  make(map[string]*FormatInfo),
	}
}

// Register adds a format to the registry.
func (r *Registry) Register(info *FormatInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.formats[info.Name]; exists {
		return fmt.Errorf("format %q already registered", info.Name)
	}

	r.formats[info.Name] = info

	for _, ext := range info.Extensions {
		ext = strings.ToLower(ext)
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		r.byExt[ext] = info
	}

	for _, mime := range info.MIMETypes {
		r.byMIME[strings.ToLower(mime)] = info
	}

	return nil
}

// Get returns format info by name.
func (r *Registry) Get(name string) (*FormatInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.formats[name]
	return info, ok
}

// GetByExtension returns format info by file extension.
func (r *Registry) GetByExtension(ext string) (*FormatInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	info, ok := r.byExt[ext]
	return info, ok
}

// GetByFilename returns format info by filename.
func (r *Registry) GetByFilename(filename string) (*FormatInfo, bool) {
	ext := filepath.Ext(filename)
	return r.GetByExtension(ext)
}

// GetByMIME returns format info by MIME type.
func (r *Registry) GetByMIME(mime string) (*FormatInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.byMIME[strings.ToLower(mime)]
	return info, ok
}

// List returns all registered formats.
func (r *Registry) List() []*FormatInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*FormatInfo, 0, len(r.formats))
	for _, info := range r.formats {
		list = append(list, info)
	}
	return list
}

// ListDecodable returns formats that support decoding.
func (r *Registry) ListDecodable() []*FormatInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*FormatInfo
	for _, info := range r.formats {
		if info.CanDecode() {
			list = append(list, info)
		}
	}
	return list
}

// ListEncodable returns formats that support encoding.
func (r *Registry) ListEncodable() []*FormatInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*FormatInfo
	for _, info := range r.formats {
		if info.CanEncode() {
			list = append(list, info)
		}
	}
	return list
}

// Register registers a format with the global registry.
func Register(info *FormatInfo) error {
	return GlobalRegistry.Register(info)
}

// init registers all built-in formats.
func init() {
	// ScreenJSON native formats
	Register(&FormatInfo{
		Name:         "json",
		Extensions:   []string{".json"},
		MIMETypes:    []string{"application/json"},
		Capabilities: CapDecode | CapEncode,
		Description:  "ScreenJSON (native JSON format)",
	})

	Register(&FormatInfo{
		Name:         "yaml",
		Extensions:   []string{".yaml", ".yml"},
		MIMETypes:    []string{"application/x-yaml", "text/yaml"},
		Capabilities: CapDecode | CapEncode,
		Description:  "ScreenYAML (native YAML format)",
	})

	// Screenplay formats
	Register(&FormatInfo{
		Name:         "fdx",
		Extensions:   []string{".fdx"},
		MIMETypes:    []string{"application/xml"},
		Capabilities: CapDecode | CapEncode,
		Description:  "Final Draft Pro (XML)",
	})

	Register(&FormatInfo{
		Name:         "fadein",
		Extensions:   []string{".fadein"},
		MIMETypes:    []string{"application/zip"},
		Capabilities: CapDecode | CapEncode,
		Description:  "FadeIn Pro (Open Screenplay Format in ZIP)",
	})

	Register(&FormatInfo{
		Name:         "fountain",
		Extensions:   []string{".fountain"},
		MIMETypes:    []string{"text/plain"},
		Capabilities: CapDecode | CapEncode,
		Description:  "Fountain (Markdown-like)",
	})

	Register(&FormatInfo{
		Name:         "celtx",
		Extensions:   []string{".celtx"},
		MIMETypes:    []string{"application/zip"},
		Capabilities: 0, // Placeholder - not yet implemented
		Description:  "Celtx (RDF in ZIP) - placeholder",
	})

	Register(&FormatInfo{
		Name:         "pdf",
		Extensions:   []string{".pdf"},
		MIMETypes:    []string{"application/pdf"},
		Capabilities: CapEncode, // Decode requires Poppler
		Description:  "PDF (export native, import requires Poppler)",
	})
}

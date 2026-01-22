// Package codec provides Celtx encoding/decoding.
// Note: Celtx support is a placeholder - full implementation requires example files.
package codec

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	celtxmodel "screenjson/cli/internal/formats/celtx/model"
	"screenjson/cli/internal/rdf"
)

// ErrNotImplemented indicates Celtx support is not yet fully implemented.
var ErrNotImplemented = fmt.Errorf("Celtx format support is not yet fully implemented")

// Decoder decodes Celtx files.
type Decoder struct{}

// NewDecoder creates a new Celtx decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Decode parses a Celtx ZIP file into the Celtx model.
// Note: This is a placeholder implementation.
func (d *Decoder) Decode(ctx context.Context, data []byte) (*celtxmodel.Project, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty Celtx input")
	}
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return nil, fmt.Errorf("Celtx input is not a ZIP archive")
	}
	// Celtx files are ZIP archives
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("Celtx ZIP open error: %w", err)
	}

	project := &celtxmodel.Project{}

	// Read project.rdf
	for _, f := range r.File {
		switch {
		case f.Name == celtxmodel.FileProjectRDF:
			data, err := readZipFile(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read project.rdf: %w", err)
			}
			project.ProjectRDF = data
			
			// Parse RDF metadata
			if err := parseProjectRDF(project, data); err != nil {
				// Non-fatal, continue
			}

		case f.Name == celtxmodel.FileLocalRDF:
			data, err := readZipFile(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read local.rdf: %w", err)
			}
			project.LocalRDF = data

		case strings.HasPrefix(f.Name, celtxmodel.ScriptPrefix) && strings.HasSuffix(f.Name, ".html"):
			// Script content
			data, err := readZipFile(f)
			if err != nil {
				continue
			}
			script := parseScriptHTML(f.Name, data)
			project.Scripts = append(project.Scripts, script)
		}
	}

	if len(project.Scripts) == 0 && len(project.ProjectRDF) == 0 {
		return nil, fmt.Errorf("no valid Celtx content found")
	}

	return project, nil
}

// readZipFile reads the contents of a zip file entry.
func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	const maxFileSize = 50 << 20 // 50MB safety limit
	limited := io.LimitReader(rc, maxFileSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > maxFileSize {
		return nil, fmt.Errorf("zip entry %s exceeds size limit (%d bytes)", f.Name, maxFileSize)
	}
	return data, nil
}

// parseProjectRDF extracts metadata from project.rdf.
func parseProjectRDF(project *celtxmodel.Project, data []byte) error {
	graph, err := rdf.ParseFile(data)
	if err != nil {
		return err
	}

	// Extract title
	titles := graph.ExtractLiterals(rdf.DCTitle.URI)
	if len(titles) > 0 {
		project.Title = titles[0]
	}

	// Extract creator/author
	creators := graph.ExtractLiterals(rdf.DCCreator.URI)
	if len(creators) > 0 {
		project.Author = creators[0]
	}

	// Extract description
	descriptions := graph.ExtractLiterals(rdf.DCDescription.URI)
	if len(descriptions) > 0 {
		project.Description = descriptions[0]
	}

	return nil
}

// parseScriptHTML parses Celtx script HTML into elements.
// Note: This is a basic implementation - Celtx HTML structure may vary.
func parseScriptHTML(filename string, data []byte) celtxmodel.Script {
	script := celtxmodel.Script{
		ID: strings.TrimSuffix(strings.TrimPrefix(filename, celtxmodel.ScriptPrefix), ".html"),
	}

	// Basic HTML parsing - in production, use proper HTML parser
	content := string(data)
	
	// Look for common Celtx HTML patterns
	// Celtx uses <p class="sceneheading">, <p class="action">, etc.
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for paragraph elements with class
		if strings.Contains(line, "<p class=\"") {
			elemType := extractClass(line)
			text := extractText(line)
			
			if elemType != "" && text != "" {
				script.Elements = append(script.Elements, celtxmodel.Element{
					Type: elemType,
					Text: text,
				})
			}
		}
	}

	return script
}

// extractClass extracts the class name from an HTML element.
func extractClass(line string) string {
	start := strings.Index(line, "class=\"")
	if start == -1 {
		return ""
	}
	start += 7
	end := strings.Index(line[start:], "\"")
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}

// extractText extracts text content from an HTML element.
func extractText(line string) string {
	// Find content between > and </p>
	start := strings.Index(line, ">")
	if start == -1 {
		return ""
	}
	end := strings.Index(line, "</p>")
	if end == -1 {
		end = len(line)
	}
	if start+1 >= end {
		return ""
	}
	
	text := line[start+1 : end]
	// Strip any remaining HTML tags
	for strings.Contains(text, "<") {
		tagStart := strings.Index(text, "<")
		tagEnd := strings.Index(text[tagStart:], ">")
		if tagEnd == -1 {
			break
		}
		text = text[:tagStart] + text[tagStart+tagEnd+1:]
	}
	
	return strings.TrimSpace(text)
}

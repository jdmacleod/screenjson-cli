package codec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"

	"screenjson/cli/internal/model"
)

// ErrPDFImportDisabled indicates PDF import requires Poppler.
var ErrPDFImportDisabled = fmt.Errorf("PDF import is disabled: pdftohtml (Poppler) not available")

// Decoder decodes PDF files using Poppler's pdftohtml.
type Decoder struct {
	pdfToHtmlPath string
}

// NewDecoder creates a new PDF decoder.
func NewDecoder(pdfToHtmlPath string) *Decoder {
	return &Decoder{
		pdfToHtmlPath: pdfToHtmlPath,
	}
}

// IsAvailable checks if PDF import is available.
func (d *Decoder) IsAvailable() bool {
	if d.pdfToHtmlPath == "" {
		return false
	}
	_, err := os.Stat(d.pdfToHtmlPath)
	return err == nil
}

// Decode parses a PDF file into a ScreenJSON document.
// This uses Poppler's pdftohtml to extract layout-aware XML.
func (d *Decoder) Decode(ctx context.Context, data []byte, password string) (*model.Document, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty PDF input")
	}
	if !d.IsAvailable() {
		return nil, ErrPDFImportDisabled
	}

	// Write PDF to temp file
	tmpPDF, err := os.CreateTemp("", "screenjson-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpPDF.Name())
	
	if _, err := tmpPDF.Write(data); err != nil {
		tmpPDF.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpPDF.Close()

	// Run pdftohtml -xml
	args := []string{"-xml", "-enc", "UTF-8", "-noframes"}
	if password != "" {
		args = append(args, "-upw", password)
	}
	args = append(args, tmpPDF.Name(), "-")

	cmd := exec.CommandContext(ctx, d.pdfToHtmlPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("pdftohtml failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("pdftohtml failed: %w", err)
	}

	// Check for OCR (minimal text content)
	if len(output) < 500 {
		return nil, fmt.Errorf("PDF appears to be OCR/image-based (insufficient text content)")
	}

	// Parse the XML output
	return parsePopplerXML(output)
}

// parsePopplerXML parses Poppler's XML output into a ScreenJSON document.
// This uses geometry-based inference to classify elements.
func parsePopplerXML(xmlData []byte) (*model.Document, error) {
	// Extract text lines with positions
	lines := extractTextLines(xmlData)
	
	if len(lines) == 0 {
		return nil, fmt.Errorf("no text content extracted from PDF")
	}

	// Classify lines based on margin clustering
	classified := classifyByMargin(lines)

	authorID := uuid.New().String()
	
	doc := &model.Document{
		ID:      uuid.New().String(),
		Version: "1.0.0",
		Generator: &model.Generator{
			Name:    "screenjson-cli",
			Version: "1.0.0",
		},
		Title:   model.Text{"en": "Imported PDF"},
		Lang:    "en",
		Charset: "utf-8",
		Dir:     "ltr",
		Authors: []model.Author{
			{
				ID:     authorID,
				Given:  "Unknown",
				Family: "Author",
			},
		},
	}

	// Extract characters
	characters := extractCharactersFromClassified(classified)
	doc.Characters = characters

	charMap := make(map[string]string)
	for _, c := range characters {
		charMap[strings.ToUpper(c.Name)] = c.ID
	}

	// Build scenes
	doc.Content = buildContent(classified, authorID, charMap)

	return doc, nil
}

// TextLine represents a line of text with position info.
type TextLine struct {
	Text   string
	Left   float64
	Top    float64
	Page   int
	Height float64
}

// ClassifiedLine is a text line with inferred element type.
type ClassifiedLine struct {
	TextLine
	Type model.ElementType
}

// extractTextLines parses Poppler XML to extract positioned text.
func extractTextLines(xmlData []byte) []TextLine {
	// Basic XML parsing - in production use proper XML parser
	var lines []TextLine
	content := string(xmlData)
	
	// Look for <text> elements with positioning
	// Format: <text top="123" left="456" ...>content</text>
	parts := strings.Split(content, "<text ")
	
	page := 1
	for _, part := range parts[1:] {
		// Check for page marker
		if strings.Contains(part, "page=\"") {
			// Extract page number
			pageIdx := strings.Index(part, "page=\"")
			if pageIdx >= 0 {
				start := pageIdx + 6
				end := strings.Index(part[start:], "\"")
				if end > 0 {
					fmt.Sscanf(part[start:start+end], "%d", &page)
				}
			}
		}
		
		var left, top, height float64
		
		// Extract left position
		if leftIdx := strings.Index(part, "left=\""); leftIdx >= 0 {
			start := leftIdx + 6
			end := strings.Index(part[start:], "\"")
			if end > 0 {
				fmt.Sscanf(part[start:start+end], "%f", &left)
			}
		}
		
		// Extract top position
		if topIdx := strings.Index(part, "top=\""); topIdx >= 0 {
			start := topIdx + 5
			end := strings.Index(part[start:], "\"")
			if end > 0 {
				fmt.Sscanf(part[start:start+end], "%f", &top)
			}
		}
		
		// Extract height
		if heightIdx := strings.Index(part, "height=\""); heightIdx >= 0 {
			start := heightIdx + 8
			end := strings.Index(part[start:], "\"")
			if end > 0 {
				fmt.Sscanf(part[start:start+end], "%f", &height)
			}
		}
		
		// Extract text content
		closeTag := strings.Index(part, ">")
		endTag := strings.Index(part, "</text>")
		if closeTag >= 0 && endTag > closeTag {
			text := strings.TrimSpace(part[closeTag+1 : endTag])
			if text != "" {
				lines = append(lines, TextLine{
					Text:   text,
					Left:   left,
					Top:    top,
					Page:   page,
					Height: height,
				})
			}
		}
	}
	
	return lines
}

// classifyByMargin classifies lines based on left margin clustering.
func classifyByMargin(lines []TextLine) []ClassifiedLine {
	if len(lines) == 0 {
		return nil
	}

	// Build histogram of left margins
	marginCounts := make(map[int]int)
	for _, line := range lines {
		margin := int(line.Left / 10) * 10 // Round to nearest 10
		marginCounts[margin]++
	}

	// Find dominant margins
	// Typical screenplay margins (px): action ~134, dialogue ~274, parenthetical ~339, character ~404
	var margins []int
	for m := range marginCounts {
		margins = append(margins, m)
	}

	// Sort and identify clusters
	// Lowest margin = action/scene headings
	// Middle margin = dialogue
	// Higher margin = parentheticals
	// Highest margin = character cues

	var classified []ClassifiedLine
	
	for _, line := range lines {
		margin := int(line.Left / 10) * 10
		text := strings.TrimSpace(line.Text)
		upper := strings.ToUpper(text)
		
		var elemType model.ElementType
		
		// Scene heading detection
		if strings.HasPrefix(upper, "INT.") || strings.HasPrefix(upper, "INT ") ||
		   strings.HasPrefix(upper, "EXT.") || strings.HasPrefix(upper, "EXT ") ||
		   strings.HasPrefix(upper, "INT/EXT") || strings.HasPrefix(upper, "I/E") {
			elemType = "scene"
		} else if strings.HasSuffix(upper, "TO:") || strings.HasSuffix(upper, "OUT.") {
			// Transition
			elemType = model.ElementTransition
		} else if margin >= 350 && text == upper && len(text) < 40 {
			// Character (high margin, all caps, short)
			elemType = model.ElementCharacter
		} else if margin >= 280 && margin < 350 && strings.HasPrefix(text, "(") {
			// Parenthetical
			elemType = model.ElementParenthetical
		} else if margin >= 200 && margin < 350 {
			// Dialogue (medium margin)
			elemType = model.ElementDialogue
		} else {
			// Action (low margin)
			elemType = model.ElementAction
		}
		
		classified = append(classified, ClassifiedLine{
			TextLine: line,
			Type:     elemType,
		})
	}
	
	return classified
}

// extractCharactersFromClassified extracts unique character names.
func extractCharactersFromClassified(lines []ClassifiedLine) []model.Character {
	seen := make(map[string]bool)
	var chars []model.Character

	for _, line := range lines {
		if line.Type == model.ElementCharacter {
			name := cleanCharacterName(line.Text)
			upperName := strings.ToUpper(name)
			
			if name != "" && !seen[upperName] {
				seen[upperName] = true
				chars = append(chars, model.Character{
					ID:   uuid.New().String(),
					Name: name,
				})
			}
		}
	}

	return chars
}

func cleanCharacterName(name string) string {
	if idx := strings.Index(name, "("); idx > 0 {
		name = name[:idx]
	}
	return strings.TrimSpace(name)
}

// buildContent builds ScreenJSON content from classified lines.
func buildContent(lines []ClassifiedLine, authorID string, charMap map[string]string) *model.Content {
	content := &model.Content{
		Cover: &model.Cover{
			Title:   model.Text{"en": "Imported PDF"},
			Authors: []string{authorID},
		},
	}

	var scenes []model.Scene
	var currentScene *model.Scene
	sceneNumber := 0

	for _, line := range lines {
		switch line.Type {
		case "scene":
			if currentScene != nil {
				scenes = append(scenes, *currentScene)
			}
			sceneNumber++
			currentScene = &model.Scene{
				ID:      uuid.New().String(),
				Authors: []string{authorID},
				Heading: parseSluglineFromText(line.Text, sceneNumber),
				Body:    []model.Element{},
			}

		default:
			if currentScene == nil {
				sceneNumber++
				currentScene = &model.Scene{
					ID:      uuid.New().String(),
					Authors: []string{authorID},
					Heading: &model.Slugline{
						No:      sceneNumber,
						Context: "INT",
						Setting: "UNKNOWN",
						Time:    "DAY",
					},
					Body: []model.Element{},
				}
			}

			elem := model.Element{
				ID:      uuid.New().String(),
				Type:    line.Type,
				Authors: []string{authorID},
			}

			switch line.Type {
			case model.ElementCharacter:
				name := cleanCharacterName(line.Text)
				elem.Character = charMap[strings.ToUpper(name)]
				elem.Display = line.Text
				if elem.Character != "" && !contains(currentScene.Cast, elem.Character) {
					currentScene.Cast = append(currentScene.Cast, elem.Character)
				}
			default:
				elem.Text = model.Text{"en": line.Text}
			}

			currentScene.Body = append(currentScene.Body, elem)
		}
	}

	if currentScene != nil {
		scenes = append(scenes, *currentScene)
	}

	content.Scenes = scenes
	return content
}

func parseSluglineFromText(text string, sceneNumber int) *model.Slugline {
	slug := &model.Slugline{
		No:   sceneNumber,
		Time: "DAY",
	}

	text = strings.TrimSpace(text)
	upper := strings.ToUpper(text)

	if strings.HasPrefix(upper, "INT/EXT") || strings.HasPrefix(upper, "I/E") {
		slug.Context = "INT/EXT"
		text = strings.TrimPrefix(upper, "INT/EXT")
		text = strings.TrimPrefix(text, "I/E")
	} else if strings.HasPrefix(upper, "INT.") || strings.HasPrefix(upper, "INT ") {
		slug.Context = "INT"
		text = strings.TrimPrefix(upper, "INT.")
		text = strings.TrimPrefix(text, "INT ")
	} else if strings.HasPrefix(upper, "EXT.") || strings.HasPrefix(upper, "EXT ") {
		slug.Context = "EXT"
		text = strings.TrimPrefix(upper, "EXT.")
		text = strings.TrimPrefix(text, "EXT ")
	} else {
		slug.Context = "INT"
	}

	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, ".")
	text = strings.TrimPrefix(text, "-")
	text = strings.TrimSpace(text)

	parts := strings.Split(text, " - ")
	if len(parts) >= 2 {
		slug.Setting = strings.TrimSpace(parts[0])
		slug.Time = strings.TrimSpace(parts[len(parts)-1])
	} else {
		slug.Setting = text
	}

	return slug
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

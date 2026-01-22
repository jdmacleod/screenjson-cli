// Package codec provides PDF encoding/decoding.
package codec

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"

	pdfmodel "screenjson/cli/internal/formats/pdf/model"
	"screenjson/cli/internal/model"
)

// Encoder encodes ScreenJSON documents to PDF.
type Encoder struct {
	setup   pdfmodel.PageSetup
	indents pdfmodel.ElementIndent
}

// NewEncoder creates a new PDF encoder with default settings.
func NewEncoder() *Encoder {
	setup := pdfmodel.DefaultPageSetup()
	return &Encoder{
		setup:   setup,
		indents: pdfmodel.DefaultIndents(),
	}
}

// WithPageSetup sets the page setup.
func (e *Encoder) WithPageSetup(setup pdfmodel.PageSetup) *Encoder {
	e.setup = setup
	return e
}

// Encode generates a PDF from a ScreenJSON document.
func (e *Encoder) Encode(ctx context.Context, doc *model.Document) ([]byte, error) {
	// Create PDF
	orientation := "P" // Portrait
	unit := "pt"       // Points
	size := "Letter"
	if e.setup.Height > 800 { // A4 is taller
		size = "A4"
	}
	
	pdf := gofpdf.New(orientation, unit, size, "")
	pdf.SetMargins(e.setup.MarginLeft, e.setup.MarginTop, e.setup.MarginRight)
	pdf.SetAutoPageBreak(true, e.setup.MarginBottom)
	
	// Set font
	pdf.SetFont("Courier", "", e.setup.FontSize)
	
	// Build character lookup
	charMap := make(map[string]string)
	for _, c := range doc.Characters {
		charMap[c.ID] = c.Name
	}
	
	// Get primary language
	lang := doc.Lang
	if lang == "" {
		lang = "en"
	}
	
	// Add first page
	pdf.AddPage()
	
	// Render cover page if we have title info
	hasCover := false
	if doc.Content != nil && doc.Content.Cover != nil {
		e.renderCoverPage(pdf, doc, lang)
		pdf.AddPage()
		hasCover = true
	}
	_ = hasCover // For future page numbering logic
	
	// Render content
	if doc.Content != nil {
		var lastCharID string
		
		for _, scene := range doc.Content.Scenes {
			// Scene heading
			if scene.Heading != nil {
				e.addBlankLine(pdf)
				heading := formatSlugline(scene.Heading)
				e.renderElement(pdf, model.ElementType("scene"), heading, "", charMap, lang)
				e.addBlankLine(pdf)
			}
			
			// Scene body
			for _, elem := range scene.Body {
				switch elem.Type {
				case model.ElementCharacter:
					e.addBlankLine(pdf)
					display := elem.Display
					if display == "" {
						if name, ok := charMap[elem.Character]; ok {
							display = strings.ToUpper(name)
						}
					}
					e.renderElement(pdf, elem.Type, display, "", charMap, lang)
					lastCharID = elem.Character
					
				case model.ElementDialogue:
					text := elem.Text.GetOrDefault(lang)
					e.renderElement(pdf, elem.Type, text, lastCharID, charMap, lang)
					
				case model.ElementParenthetical:
					text := elem.Text.GetOrDefault(lang)
					if !strings.HasPrefix(text, "(") {
						text = "(" + text
					}
					if !strings.HasSuffix(text, ")") {
						text = text + ")"
					}
					e.renderElement(pdf, elem.Type, text, "", charMap, lang)
					
				case model.ElementAction:
					e.addBlankLine(pdf)
					text := elem.Text.GetOrDefault(lang)
					e.renderElement(pdf, elem.Type, text, "", charMap, lang)
					
				case model.ElementTransition:
					e.addBlankLine(pdf)
					text := strings.ToUpper(elem.Text.GetOrDefault(lang))
					e.renderElement(pdf, elem.Type, text, "", charMap, lang)
					
				case model.ElementShot:
					e.addBlankLine(pdf)
					text := strings.ToUpper(elem.Text.GetOrDefault(lang))
					e.renderElement(pdf, elem.Type, text, "", charMap, lang)
					
				case model.ElementGeneral:
					text := elem.Text.GetOrDefault(lang)
					e.renderElement(pdf, elem.Type, text, "", charMap, lang)
				}
			}
		}
	}
	
	// Add page numbers (starting from page 2)
	totalPages := pdf.PageCount()
	for i := 2; i <= totalPages; i++ {
		pdf.SetPage(i)
		pdf.SetY(36) // 0.5 inch from top
		pdf.SetX(e.setup.Width - e.setup.MarginRight - 30)
		pdf.SetFont("Courier", "", 12)
		pdf.CellFormat(30, 12, fmt.Sprintf("%d.", i-1), "", 0, "R", false, 0, "")
	}
	
	// Generate output
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("PDF generation error: %w", err)
	}

	return buf.Bytes(), nil
}

// renderCoverPage renders the title page.
func (e *Encoder) renderCoverPage(pdf *gofpdf.Fpdf, doc *model.Document, lang string) {
	pdf.SetFont("Courier", "B", 24)
	
	// Center the title
	title := doc.Title.GetOrDefault(lang)
	pdf.SetY(e.setup.Height / 3)
	pdf.SetX(e.setup.MarginLeft)
	pdf.MultiCell(e.setup.TextWidth(), 30, strings.ToUpper(title), "", "C", false)
	
	// Author(s)
	pdf.SetFont("Courier", "", 12)
	pdf.Ln(36)
	
	if len(doc.Authors) > 0 {
		pdf.SetX(e.setup.MarginLeft)
		pdf.MultiCell(e.setup.TextWidth(), 14, "Written by", "", "C", false)
		
		for _, author := range doc.Authors {
			name := strings.TrimSpace(author.Given + " " + author.Family)
			pdf.SetX(e.setup.MarginLeft)
			pdf.MultiCell(e.setup.TextWidth(), 14, name, "", "C", false)
		}
	}
	
	// Source material
	if len(doc.Sources) > 0 {
		pdf.Ln(24)
		for _, source := range doc.Sources {
			sourceTitle := source.Title.GetOrDefault(lang)
			if sourceTitle != "" {
				pdf.SetX(e.setup.MarginLeft)
				text := fmt.Sprintf("Based on %s \"%s\"", source.Type, sourceTitle)
				pdf.MultiCell(e.setup.TextWidth(), 14, text, "", "C", false)
			}
		}
	}
}

// renderElement renders a single screenplay element.
func (e *Encoder) renderElement(pdf *gofpdf.Fpdf, elemType model.ElementType, text, charID string, charMap map[string]string, lang string) {
	// Get current Y position
	y := pdf.GetY()
	
	// Calculate indentation and width based on element type
	var indent, width float64
	align := "L"
	
	switch elemType {
	case "scene":
		indent = 0
		width = e.setup.TextWidth()
		text = strings.ToUpper(text)
		
	case model.ElementAction:
		indent = e.indents.Action
		width = e.setup.TextWidth()
		
	case model.ElementCharacter:
		indent = e.indents.Character
		width = 200
		text = strings.ToUpper(text)
		
	case model.ElementDialogue:
		indent = e.indents.Dialogue
		width = 252 // About 3.5 inches
		
	case model.ElementParenthetical:
		indent = e.indents.Parenthetical
		width = 180
		
	case model.ElementTransition:
		// Right-aligned
		align = "R"
		indent = 0
		width = e.setup.TextWidth()
		text = strings.ToUpper(text)
		
	case model.ElementShot:
		indent = 0
		width = e.setup.TextWidth()
		text = strings.ToUpper(text)
		
	default:
		indent = 0
		width = e.setup.TextWidth()
	}
	
	// Set position
	pdf.SetY(y)
	pdf.SetX(e.setup.MarginLeft + indent)
	
	// Render text with word wrap
	pdf.MultiCell(width, e.setup.LineHeight, text, "", align, false)
}

// addBlankLine adds a blank line.
func (e *Encoder) addBlankLine(pdf *gofpdf.Fpdf) {
	pdf.Ln(e.setup.LineHeight)
}

// formatSlugline formats a scene heading.
func formatSlugline(slug *model.Slugline) string {
	var parts []string
	
	context := slug.Context
	if context == "" {
		context = "INT"
	}
	parts = append(parts, context)
	
	if slug.Setting != "" {
		parts = append(parts, slug.Setting)
	}
	
	result := strings.Join(parts, ". ")
	
	if slug.Time != "" {
		result += " - " + slug.Time
	}
	
	return strings.ToUpper(result)
}

// Package model defines PDF layout and rendering structures.
package model

// PageSetup defines industry-standard screenplay page layout.
type PageSetup struct {
	// Page dimensions (in points, 72 points = 1 inch)
	Width  float64 // US Letter: 612
	Height float64 // US Letter: 792
	
	// Margins (in points)
	MarginTop    float64 // 72 (1 inch)
	MarginBottom float64 // 72 (1 inch)
	MarginLeft   float64 // 108 (1.5 inches)
	MarginRight  float64 // 72 (1 inch)
	
	// Font settings
	FontFamily string  // Courier or Courier Prime
	FontSize   float64 // 12 pt
	LineHeight float64 // typically 12 pt for single spacing
	
	// Characters per inch (monospace)
	CharsPerInch float64 // 10 for Courier
}

// DefaultPageSetup returns industry-standard screenplay page setup.
func DefaultPageSetup() PageSetup {
	return PageSetup{
		Width:        612,  // US Letter width
		Height:       792,  // US Letter height
		MarginTop:    72,   // 1 inch
		MarginBottom: 72,   // 1 inch
		MarginLeft:   108,  // 1.5 inches
		MarginRight:  72,   // 1 inch
		FontFamily:   "Courier",
		FontSize:     12,
		LineHeight:   12,
		CharsPerInch: 10,
	}
}

// A4PageSetup returns A4 page setup.
func A4PageSetup() PageSetup {
	setup := DefaultPageSetup()
	setup.Width = 595.28   // A4 width
	setup.Height = 841.89  // A4 height
	return setup
}

// TextWidth returns the usable text width.
func (p PageSetup) TextWidth() float64 {
	return p.Width - p.MarginLeft - p.MarginRight
}

// TextHeight returns the usable text height.
func (p PageSetup) TextHeight() float64 {
	return p.Height - p.MarginTop - p.MarginBottom
}

// LinesPerPage returns approximate lines per page.
func (p PageSetup) LinesPerPage() int {
	return int(p.TextHeight() / p.LineHeight)
}

// ElementIndent defines indentation for each element type (in points).
type ElementIndent struct {
	Action        float64 // 0 (left margin)
	Dialogue      float64 // 180 (2.5 inches from left)
	Parenthetical float64 // 216 (3.0 inches from left)
	Character     float64 // 266 (3.7 inches from left)
	Transition    float64 // Right-aligned
	Shot          float64 // 0 (like action)
}

// DefaultIndents returns industry-standard indentation.
func DefaultIndents() ElementIndent {
	return ElementIndent{
		Action:        0,
		Dialogue:      180,  // 2.5 inches
		Parenthetical: 216,  // 3.0 inches
		Character:     266,  // 3.7 inches
		Transition:    0,    // Right-aligned (handled specially)
		Shot:          0,
	}
}

// ElementWidth defines max width for each element type (in points).
type ElementWidth struct {
	Action        float64 // Full width
	Dialogue      float64 // Narrower (about 3.5 inches)
	Parenthetical float64 // Even narrower (about 2.5 inches)
	Character     float64 // Short (name only)
	Transition    float64 // Short (right-aligned)
}

// DefaultWidths returns default element widths.
func DefaultWidths(setup PageSetup) ElementWidth {
	return ElementWidth{
		Action:        setup.TextWidth(),
		Dialogue:      252,  // About 3.5 inches
		Parenthetical: 180,  // About 2.5 inches
		Character:     200,  // Enough for names
		Transition:    150,  // Short
	}
}

// Package model defines Celtx format structures.
// Celtx files are ZIP archives containing RDF/XML metadata and HTML script content.
package model

// Project represents a Celtx project structure.
type Project struct {
	// RDF metadata
	Title       string
	Author      string
	Description string
	Created     string
	Modified    string
	
	// Script content (extracted from HTML)
	Scripts []Script
	
	// Raw RDF data
	ProjectRDF []byte
	LocalRDF   []byte
}

// Script represents a script within a Celtx project.
type Script struct {
	ID       string
	Title    string
	Elements []Element
}

// Element represents a screenplay element in Celtx HTML.
type Element struct {
	Type string // sceneheading, action, character, dialog, parenthetical, transition
	Text string
}

// Celtx element type constants (as used in HTML class names)
const (
	TypeSceneHeading   = "sceneheading"
	TypeAction         = "action"
	TypeCharacter      = "character"
	TypeDialog         = "dialog"
	TypeParenthetical  = "parenthetical"
	TypeTransition     = "transition"
)

// Celtx file names within the ZIP archive
const (
	FileProjectRDF = "project.rdf"
	FileLocalRDF   = "local.rdf"
	ScriptPrefix   = "script-"
	ScratchPrefix  = "scratch-"
)

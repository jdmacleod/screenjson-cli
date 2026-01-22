package rdf

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Parser parses RDF/XML documents.
type Parser struct {
	namespaces map[string]string
}

// NewParser creates a new RDF/XML parser.
func NewParser() *Parser {
	return &Parser{
		namespaces: make(map[string]string),
	}
}

// Parse parses RDF/XML data into a Graph.
func (p *Parser) Parse(data []byte) (*Graph, error) {
	graph := NewGraph()
	
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	
	var currentSubject string
	var elementStack []string
	
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML parse error: %w", err)
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			name := t.Name.Local
			ns := t.Name.Space
			
			// Check for rdf:RDF root
			if name == "RDF" && strings.Contains(ns, "rdf") {
				// Parse namespace declarations
				for _, attr := range t.Attr {
					if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" {
						p.namespaces[attr.Name.Local] = attr.Value
					}
				}
				continue
			}
			
			// Check for rdf:Description or typed node
			if name == "Description" || len(elementStack) == 0 {
				// Get subject from rdf:about or rdf:ID
				for _, attr := range t.Attr {
					if attr.Name.Local == "about" || attr.Name.Local == "ID" {
						currentSubject = attr.Value
						break
					}
				}
				
				// If typed node (not Description), add type triple
				if name != "Description" && currentSubject != "" {
					typeURI := ns + name
					graph.Add(Triple{
						Subject:   Resource{URI: currentSubject},
						Predicate: RDFtype,
						Object:    Resource{URI: typeURI},
					})
				}
				
				// Process attributes as property values
				for _, attr := range t.Attr {
					if attr.Name.Local != "about" && attr.Name.Local != "ID" && 
					   attr.Name.Space != "xmlns" && attr.Name.Local != "xmlns" {
						predURI := attr.Name.Space + attr.Name.Local
						graph.Add(Triple{
							Subject:   Resource{URI: currentSubject},
							Predicate: Resource{URI: predURI},
							Object:    Literal{Value: attr.Value},
						})
					}
				}
			} else if currentSubject != "" {
				// Property element
				elementStack = append(elementStack, ns+name)
			}
			
		case xml.EndElement:
			if len(elementStack) > 0 {
				elementStack = elementStack[:len(elementStack)-1]
			}
			
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(elementStack) > 0 && currentSubject != "" {
				predURI := elementStack[len(elementStack)-1]
				graph.Add(Triple{
					Subject:   Resource{URI: currentSubject},
					Predicate: Resource{URI: predURI},
					Object:    Literal{Value: text},
				})
			}
		}
	}
	
	return graph, nil
}

// ParseFile parses an RDF/XML file.
func ParseFile(data []byte) (*Graph, error) {
	p := NewParser()
	return p.Parse(data)
}

// ExtractLiterals extracts all literal values for a given predicate.
func (g *Graph) ExtractLiterals(predicate string) []string {
	var values []string
	for _, t := range g.Triples {
		if t.Predicate.URI == predicate {
			if lit, ok := t.Object.(Literal); ok {
				values = append(values, lit.Value)
			}
		}
	}
	return values
}

// GetSubjects returns all unique subjects in the graph.
func (g *Graph) GetSubjects() []string {
	seen := make(map[string]bool)
	var subjects []string
	for _, t := range g.Triples {
		if !seen[t.Subject.URI] {
			seen[t.Subject.URI] = true
			subjects = append(subjects, t.Subject.URI)
		}
	}
	return subjects
}

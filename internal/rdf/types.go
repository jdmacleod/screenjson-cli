// Package rdf provides basic RDF/XML parsing for Celtx format support.
package rdf

// Triple represents an RDF triple (subject, predicate, object).
type Triple struct {
	Subject   Resource
	Predicate Resource
	Object    Node
}

// Resource represents an RDF resource (URI).
type Resource struct {
	URI string
}

// Node represents an RDF node (resource or literal).
type Node interface {
	IsResource() bool
	IsLiteral() bool
	String() string
}

// Literal represents an RDF literal value.
type Literal struct {
	Value    string
	Language string // e.g., "en"
	Datatype string // e.g., "xsd:string"
}

// IsResource returns false for Literal.
func (l Literal) IsResource() bool { return false }

// IsLiteral returns true for Literal.
func (l Literal) IsLiteral() bool { return true }

// String returns the literal value.
func (l Literal) String() string { return l.Value }

// IsResource returns true for Resource.
func (r Resource) IsResource() bool { return true }

// IsLiteral returns false for Resource.
func (r Resource) IsLiteral() bool { return false }

// String returns the resource URI.
func (r Resource) String() string { return r.URI }

// Graph represents a collection of RDF triples.
type Graph struct {
	Triples []Triple
	Base    string
}

// NewGraph creates a new empty RDF graph.
func NewGraph() *Graph {
	return &Graph{
		Triples: make([]Triple, 0),
	}
}

// Add adds a triple to the graph.
func (g *Graph) Add(t Triple) {
	g.Triples = append(g.Triples, t)
}

// FindBySubject returns all triples with the given subject.
func (g *Graph) FindBySubject(subject string) []Triple {
	var results []Triple
	for _, t := range g.Triples {
		if t.Subject.URI == subject {
			results = append(results, t)
		}
	}
	return results
}

// FindByPredicate returns all triples with the given predicate.
func (g *Graph) FindByPredicate(predicate string) []Triple {
	var results []Triple
	for _, t := range g.Triples {
		if t.Predicate.URI == predicate {
			results = append(results, t)
		}
	}
	return results
}

// FindBySubjectPredicate returns objects matching subject and predicate.
func (g *Graph) FindBySubjectPredicate(subject, predicate string) []Node {
	var results []Node
	for _, t := range g.Triples {
		if t.Subject.URI == subject && t.Predicate.URI == predicate {
			results = append(results, t.Object)
		}
	}
	return results
}

// Common RDF namespace URIs
const (
	NSrdf  = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	NSrdfs = "http://www.w3.org/2000/01/rdf-schema#"
	NSxsd  = "http://www.w3.org/2001/XMLSchema#"
	NSdc   = "http://purl.org/dc/elements/1.1/"
	NSdcterms = "http://purl.org/dc/terms/"
)

// Common RDF predicates
var (
	RDFtype    = Resource{URI: NSrdf + "type"}
	RDFSLabel  = Resource{URI: NSrdfs + "label"}
	DCTitle    = Resource{URI: NSdc + "title"}
	DCCreator  = Resource{URI: NSdc + "creator"}
	DCDate     = Resource{URI: NSdc + "date"}
	DCDescription = Resource{URI: NSdc + "description"}
)

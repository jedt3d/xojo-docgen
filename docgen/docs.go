package main

import (
	"strings"
)

// docs.go holds the documentation-extraction helpers shared between parsing
// and rendering. Most extraction happens inline in parser.go (because Notes
// are children of their parent #tag block); this file provides precedence
// helpers and formatted descriptions.

// entityDocs renders the full description text for a member/container given
// its collected notes, named notes, and leading comments. Precedence:
//
//	Notes (unnamed) > LibraryDescription attribute > leading ' comments.
func entityDocs(notes string, named []NamedNote, comments []string, attrs []Attribute) string {
	var b strings.Builder
	// LibraryDescription may carry the canonical description.
	libDesc := ""
	for _, a := range attrs {
		if strings.EqualFold(a.Name, "LibraryDescription") && a.Value != "" {
			libDesc = a.Value
		}
	}
	if strings.TrimSpace(notes) != "" {
		b.WriteString(strings.TrimSpace(notes))
	} else if libDesc != "" {
		b.WriteString(libDesc)
	} else if len(comments) > 0 {
		b.WriteString(strings.Join(comments, "\n"))
	}
	return strings.TrimRight(b.String(), "\n")
}

// memberDocs returns the docs text for any Member.
func memberDocs(m Member) string {
	switch v := m.(type) {
	case *Method:
		return entityDocs(v.Notes, v.NamedNotes, v.DocComments, v.Attributes)
	case *Property:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	case *ComputedProperty:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	case *Constant:
		return entityDocs("", nil, nil, nil) // constants: no body docs in this fixture
	case *Enum:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	case *Delegate:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	case *EventDef:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	case *EventHandler:
		return entityDocs(v.Notes, nil, v.DocComments, nil)
	}
	return ""
}

// containerDocs returns the docs text for a Container.
func containerDocs(c *Container) string {
	return entityDocs(c.Notes, c.NamedNotes, c.DocComments, c.Attributes)
}

// isDeprecated reports whether a member/container is deprecated.
func memberDeprecated(m Member) (bool, string) {
	switch v := m.(type) {
	case *Method:
		return v.IsDeprecated, v.DeprecMsg
	case *Property:
		return false, ""
	case *ComputedProperty:
		return false, ""
	}
	return false, ""
}

// containerDeprecated scans a container's attributes for a Deprecated entry.
func containerDeprecated(c *Container) (bool, string) {
	for _, a := range c.Attributes {
		if strings.EqualFold(a.Name, "Deprecated") {
			return true, a.Value
		}
	}
	return false, ""
}

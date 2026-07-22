package main

import (
	"regexp"
	"strings"
)

// parseSubFunctionLine parses the first body line of a Method/Event/Delegate:
//
//	Sub DoThing()
//	Sub DoThing(name As String)
//	Function Add(x As Integer, y As Integer) As Integer
//	Private Sub Foo()           <- the "Private"/"Public"/"Protected" prefix is sometimes present
//	Function GetYears() As String()
//
// Returns the parsed method shape and the raw declaration line.
func parseSubFunctionLine(line string) (name string, isFunction bool, params []Param, returnType string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false, nil, "", false
	}
	// Strip leading scope/shared keyword if present.
	for _, kw := range []string{"Private ", "Public ", "Protected ", "Shared "} {
		if strings.HasPrefix(line, kw) {
			line = strings.TrimSpace(strings.TrimPrefix(line, kw))
			break
		}
	}
	// Must start with Sub or Function.
	upperHead := line
	if strings.HasPrefix(upperHead, "Sub ") {
		isFunction = false
		upperHead = strings.TrimPrefix(upperHead, "Sub ")
	} else if strings.HasPrefix(upperHead, "Function ") {
		isFunction = true
		upperHead = strings.TrimPrefix(upperHead, "Function ")
	} else {
		return "", false, nil, "", false
	}
	upperHead = strings.TrimSpace(upperHead)

	// Split into name(params) [As Return]
	// Find the first '(' and its matching ')'.
	open := strings.IndexByte(upperHead, '(')
	if open < 0 {
		// No params: "Sub Foo" or "Function Foo As Integer"
		// Name is up to " As " or end.
		if asIdx := strings.Index(upperHead, " As "); asIdx >= 0 {
			name = strings.TrimSpace(upperHead[:asIdx])
			returnType = strings.TrimSpace(upperHead[asIdx+4:])
		} else {
			name = upperHead
		}
		// strip trailing "End Sub"/etc just in case (shouldn't be here)
		name = strings.TrimSpace(strings.TrimSuffix(name, "End"))
		return name, isFunction, nil, returnType, true
	}
	close := matchingParen(upperHead, open)
	if close < 0 {
		// malformed; fall back
		name = strings.TrimSpace(upperHead[:open])
		return name, isFunction, nil, "", true
	}
	name = strings.TrimSpace(upperHead[:open])
	paramStr := upperHead[open+1 : close]
	rest := strings.TrimSpace(upperHead[close+1:])
	if strings.HasPrefix(rest, "As ") {
		returnType = strings.TrimSpace(rest[3:])
		// Some signatures have trailing "Handles Foo.Bar" or "End Function" — strip those.
		for _, sep := range []string{" Handles ", " End "} {
			if i := strings.Index(returnType, sep); i >= 0 {
				returnType = strings.TrimSpace(returnType[:i])
			}
		}
	}
	params = parseParamList(paramStr)
	return name, isFunction, params, returnType, true
}

// matchingParen returns the index of the ')' that closes the '(' at openIdx,
// respecting nesting. Returns -1 if not found.
func matchingParen(s string, openIdx int) int {
	if openIdx < 0 || openIdx >= len(s) || s[openIdx] != '(' {
		return -1
	}
	depth := 0
	inQ := false
	for i := openIdx; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQ = !inQ
			continue
		}
		if inQ {
			continue
		}
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// parseParamList splits "(a As Integer, b As String)" into params, respecting
// nested parens (e.g. array types, or default values containing commas).
func parseParamList(s string) []Param {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var params []Param
	chunks := splitParams(s)
	for _, ch := range chunks {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		params = append(params, parseParam(ch))
	}
	return params
}

// splitParams splits a param list on top-level commas (respecting parens and quotes).
func splitParams(s string) []string {
	var out []string
	var cur strings.Builder
	depth := 0
	inQ := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQ = !inQ
			cur.WriteByte(c)
			continue
		}
		if inQ {
			cur.WriteByte(c)
			continue
		}
		switch c {
		case '(', '[':
			depth++
			cur.WriteByte(c)
		case ')', ']':
			depth--
			cur.WriteByte(c)
		case ',':
			if depth == 0 {
				out = append(out, cur.String())
				cur.Reset()
			} else {
				cur.WriteByte(c)
			}
		default:
			cur.WriteByte(c)
		}
	}
	out = append(out, cur.String())
	return out
}

// parseParam parses one parameter like "ByRef name As String" or
// "Optional limit As Integer = 10".
func parseParam(s string) Param {
	p := Param{Raw: strings.TrimSpace(s)}
	rest := strings.TrimSpace(s)
	// Modifiers may stack at the front.
	for {
		trimmed := strings.TrimSpace(rest)
		matched := false
		for _, mod := range []string{"ByRef ", "ByVal ", "Optional ", "ParamArray ", "Assigns ", "Extends "} {
			if hasPrefixFold(trimmed, mod) {
				rest = strings.TrimSpace(trimmed[len(mod):])
				switch {
				case mod == "ByRef ":
					p.ByRef = true
				case mod == "ByVal ":
					p.ByVal = true
				case mod == "Optional ":
					p.Optional = true
				case mod == "ParamArray ":
					p.ParamArray = true
				case mod == "Assigns ":
					p.Assigns = true
				case mod == "Extends ":
					p.Extends = true
				}
				matched = true
				break
			}
		}
		if !matched {
			break
		}
	}
	// Now rest is "name As Type [= default]" or just "name".
	// Name is the first token.
	sp := strings.IndexByte(rest, ' ')
	if sp < 0 {
		p.Name = rest
		return p
	}
	p.Name = rest[:sp]
	tail := strings.TrimSpace(rest[sp:])
	// Optional default value.
	defaultVal := ""
	if eq := indexEqualsOutsideQuotes(tail); eq >= 0 {
		defaultVal = strings.TrimSpace(tail[eq+1:])
		tail = strings.TrimSpace(tail[:eq])
	}
	if hasPrefixFold(tail, "As ") {
		p.Type = strings.TrimSpace(tail[3:])
	}
	p.Default = stripQuotes(defaultVal)
	return p
}

// hasPrefixFold reports whether s starts with prefix under ASCII case-insensitive compare.
func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, b := s[i], prefix[i]
		if a >= 'A' && a <= 'Z' {
			a += 32
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		if a != b {
			return false
		}
	}
	return true
}

// indexEqualsOutsideQuotes returns the index of the first '=' outside double quotes, or -1.
func indexEqualsOutsideQuotes(s string) int {
	inQ := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQ = !inQ
			continue
		}
		if c == '=' && !inQ {
			return i
		}
	}
	return -1
}

// propertyDeclRe matches a property declaration "Name As Type [= default]" or
// "Private Name As Type". It captures name, type, default.
var propertyDeclRe = regexp.MustCompile(`^\s*(?:(Private|Public|Protected)\s+)?([A-Za-z_][A-Za-z0-9_]*)\s+As\s+([^=\n]+?)(?:\s*=\s*(.+))?\s*$`)

// parsePropertyDecl parses "CustomerName As String" or "Count As Integer = 0".
// Returns (name, type, default, scopeKeyword, ok).
func parsePropertyDecl(line string) (name, typ, def, scopeKw string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", "", "", false
	}
	m := propertyDeclRe.FindStringSubmatch(line)
	if m == nil {
		return "", "", "", "", false
	}
	scopeKw = m[1]
	name = m[2]
	typ = strings.TrimSpace(m[3])
	if m[4] != "" {
		def = stripQuotes(strings.TrimSpace(m[4]))
	}
	return name, typ, def, scopeKw, true
}

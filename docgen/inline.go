package main

import "strings"

// parseInlineProps parses the comma-separated Key=Value properties that follow
// a #tag name, e.g.  Method, Flags = &h0, Scope = Public
// or a multi-attribute tag line like:
//
//	Constant, Name = About, Type = String, Dynamic = True, Default = "5", Scope = Protected
//
// The hard part: commas can appear inside quoted values (Default = "Hello, world")
// and inside hex IDs. The scanner respects double-quotes.
//
// The input is the text after the tag keyword, e.g.  ", Flags = &h0, Name = Foo".
func parseInlineProps(s string) map[string]string {
	out := map[string]string{}
	// Strip a leading comma if present (Xojo writes ", Flags = ...").
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, ",")
	s = strings.TrimSpace(s)
	if s == "" {
		return out
	}
	// Split into key=value chunks on commas that are outside quotes.
	chunks := splitOutsideQuotes(s, ',')
	for _, ch := range chunks {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		eq := strings.IndexByte(ch, '=')
		if eq < 0 {
			// bare flag with no value — store as true-ish
			out[ch] = ""
			continue
		}
		key := strings.TrimSpace(ch[:eq])
		val := strings.TrimSpace(ch[eq+1:])
		val = stripQuotes(val)
		out[key] = val
	}
	return out
}

// splitOutsideQuotes splits s on sep, ignoring separators that appear inside double quotes.
func splitOutsideQuotes(s string, sep byte) []string {
	var out []string
	var cur strings.Builder
	inQ := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQ = !inQ
			cur.WriteByte(c)
			continue
		}
		if c == sep && !inQ {
			out = append(out, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	out = append(out, cur.String())
	return out
}

// stripQuotes removes one matched pair of surrounding double quotes from a value.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// parseKeyValueLine parses a "Key = Value" line from a Begin/End block (or
// #tag Session block). Returns ok=false if there's no '='. Respects quotes
// in the value but does not split on commas (these are single key=value lines).
func parseKeyValueLine(line string) (key, val string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:eq])
	val = strings.TrimSpace(line[eq+1:])
	// Align-style Begin blocks sometimes pad: "Caption   =   "OK"" -> already trimmed.
	val = stripQuotes(val)
	return key, val, true
}

// unescapeConstantValue reverses the escaping used in #tag Constant Default= values.
// Xojo encodes \r \n \" \\ \x2C (comma) etc. We do a light unescape for display.
func unescapeConstantValue(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'r':
			b.WriteByte('\r')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case '"':
			b.WriteByte('"')
		case '\\':
			b.WriteByte('\\')
		case 'x':
			// \xNN — best effort, take up to 2 hex digits
			hex := ""
			for j := 0; j < 2 && i+1 < len(s); j++ {
				i++
				c := s[i]
				if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
					hex += string(c)
				} else {
					i--
					break
				}
			}
			if hex != "" {
				// interpret
				var v int
				for _, h := range hex {
					v *= 16
					switch {
					case h >= '0' && h <= '9':
						v += int(h - '0')
					case h >= 'a' && h <= 'f':
						v += int(h-'a') + 10
					case h >= 'A' && h <= 'F':
						v += int(h-'A') + 10
					}
				}
				b.WriteByte(byte(v))
			}
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

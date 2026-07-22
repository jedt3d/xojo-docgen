package main

import (
	"bufio"
	"os"
	"strings"
)

// parseFile parses one .xojo_code / .xojo_window / .xojo_menu / .xojo_toolbar
// file and populates the given container's members, notes, attributes, and
// (for pages/windows) the control tree.
func parseFile(path string, c *Container) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	var lines []string
	for sc.Scan() {
		lines = append(lines, strings.TrimRight(sc.Text(), "\r"))
	}
	if err := sc.Err(); err != nil {
		return err
	}
	p := newTagParser(lines, c)
	p.run()
	return nil
}

// tagParser is a line-based scanner over a single file.
type tagParser struct {
	lines     []string
	i         int // current line index
	container *Container
}

func newTagParser(lines []string, c *Container) *tagParser {
	return &tagParser{lines: lines, container: c}
}

func (p *tagParser) cur() string {
	if p.i < len(p.lines) {
		return p.lines[p.i]
	}
	return ""
}

func (p *tagParser) hasNext() bool { return p.i < len(p.lines) }
func (p *tagParser) advance()      { p.i++ }

// run drives the top-level parse. It expects a single outer #tag block.
func (p *tagParser) run() {
	// Find the outer #tag block opening (e.g. "#tag Class", "#tag WebPage").
	for p.hasNext() {
		line := strings.TrimSpace(p.cur())
		if strings.HasPrefix(line, "#tag ") || strings.HasPrefix(line, "#Tag ") {
			break
		}
		p.advance()
	}
	if !p.hasNext() {
		return
	}
	// Parse the outer block.
	p.parseOuterBlock()
}

// parseOuterBlock consumes the outer #tag Class/Module/WebPage/... block,
// including parsing the class header line, any Inherits/Implements, and the
// member blocks inside until the matching #tag End<Class/...>.
func (p *tagParser) parseOuterBlock() {
	header := p.cur() // "#tag Class", "#tag Module", "#tag WebPage", etc.
	tagName, _ := parseTagName(header)
	p.advance()

	// For Class/Module/Interface, the next non-blank line is the declaration.
	if tagName == "Class" || tagName == "Module" || tagName == "Interface" {
		p.parseTypeHeader(tagName)
	}

	// Now parse body members until #tag End<Name>.
	endTag := "End" + tagName
	for p.hasNext() {
		raw := p.cur()
		line := strings.TrimSpace(raw)
		if isEndTag(line, tagName) || strings.EqualFold(line, "#tag "+endTag) {
			p.advance()
			return
		}
		// The body can contain:
		//  - member #tag blocks (Method, Property, ...)
		//  - #tag WindowCode / #tag Events <Name> (for pages)
		//  - #tag ViewBehavior (skip)
		//  - #tag Session (Session class settings — skip content)
		//  - #tag Note (class-level note)
		//  - Begin/End UI trees (for pages/menus/toolbars)
		//  - property lines for menus/toolbars Begin blocks
		if strings.HasPrefix(line, "#tag ") || strings.HasPrefix(line, "#Tag ") {
			p.parseMemberOrSection(tagName)
			continue
		}
		if strings.HasPrefix(line, "Begin ") {
			// UI control / designer block at top level (page/menu/toolbar).
			ctrl := p.parseBeginBlock()
			if ctrl != nil {
				p.container.Controls = append(p.container.Controls, ctrl)
			}
			continue
		}
		// Otherwise it's a stray line inside the type body (e.g. a comment,
		// or "Inherits"/"Implements" already handled). Skip.
		p.advance()
	}
}

// parseTypeHeader reads the "(Protected) Class Name" line and optional
// "Inherits X" / "Implements Y" lines for a Class/Module/Interface.
func (p *tagParser) parseTypeHeader(tagName string) {
	// Skip blanks.
	for p.hasNext() && strings.TrimSpace(p.cur()) == "" {
		p.advance()
	}
	if !p.hasNext() {
		return
	}
	line := strings.TrimSpace(p.cur())
	// Expect "<scope?> Class/Module/Interface Name".
	scopeKw := ""
	for _, kw := range []string{"Private ", "Public ", "Protected "} {
		if strings.HasPrefix(line, kw) {
			scopeKw = strings.TrimSpace(kw)
			line = strings.TrimSpace(strings.TrimPrefix(line, kw))
			break
		}
	}
	p.container.ScopeKw = scopeKw
	if scopeKw != "" {
		p.container.Scope = scopeFromString(scopeKw)
	}
	// Drop the "Class"/"Module"/"Interface" keyword.
	for _, kw := range []string{"Class ", "Module ", "Interface "} {
		if strings.HasPrefix(line, kw) {
			line = strings.TrimSpace(strings.TrimPrefix(line, kw))
			break
		}
	}
	// What remains (up to any whitespace) is the name.
	if name := strings.TrimSpace(line); name != "" {
		// The container name from the manifest is authoritative; keep it.
		_ = name
	}
	p.advance()

	// Now look for Inherits/Implements/Aggregates lines.
	for p.hasNext() {
		l := strings.TrimSpace(p.cur())
		if l == "" {
			p.advance()
			continue
		}
		switch {
		case strings.HasPrefix(l, "Inherits "):
			p.container.Super = strings.TrimSpace(strings.TrimPrefix(l, "Inherits "))
			p.advance()
		case strings.HasPrefix(l, "Implements "):
			rest := strings.TrimSpace(strings.TrimPrefix(l, "Implements "))
			p.container.Implements = append(p.container.Implements, splitInterfaces(rest)...)
			p.advance()
		case strings.HasPrefix(l, "Aggregates "):
			rest := strings.TrimSpace(strings.TrimPrefix(l, "Aggregates "))
			p.container.Implements = append(p.container.Implements, splitInterfaces(rest)...)
			p.advance()
		default:
			// reached the first member or #tag block — stop.
			return
		}
	}
}

// parseMemberOrSection handles a #tag line that begins a member or section
// inside the outer block.
func (p *tagParser) parseMemberOrSection(outer string) {
	header := p.cur()
	tagName, inline := parseTagName(header)
	switch tagName {
	case "Method":
		p.advance()
		m := p.parseMethodBody(inline)
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "Event":
		p.advance()
		m := p.parseEventHandlerBody(inline, "")
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "Hook":
		p.advance()
		m := p.parseEventDefBody(inline)
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "Delegate":
		p.advance()
		m := p.parseDelegateBody(inline)
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "Property":
		p.advance()
		m := p.parsePropertyBody(inline)
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "ComputedProperty":
		p.advance()
		m := p.parseComputedPropertyBody(inline)
		if m != nil {
			p.container.Members = append(p.container.Members, m)
		}
	case "Constant":
		p.advance()
		p.parseConstantBlock(inline)
	case "Enum":
		p.advance()
		p.parseEnumBlock(inline)
	case "Note":
		p.advance()
		body := p.collectUntilEnd("Note")
		props := parseInlineProps(inline)
		if name, ok := props["Name"]; ok {
			p.container.NamedNotes = append(p.container.NamedNotes, NamedNote{Name: name, Body: body})
		} else {
			if p.container.Notes != "" {
				p.container.Notes += "\n\n"
			}
			p.container.Notes += body
		}
	case "Attribute":
		p.advance()
		body := p.collectUntilEnd("Attribute")
		if a := parseAttributeBody(body); a != nil {
			p.container.Attributes = append(p.container.Attributes, *a)
		}
	case "WindowCode":
		p.advance()
		p.parseWindowCodeSection()
	case "Events":
		// #tag Events <ControlName>
		p.advance()
		ctrlName := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(header, "#tag Events"), "#Tag Events"))
		p.parseEventsSection(ctrlName)
	case "ViewBehavior", "Session", "MenuHandler":
		// Skip these sections entirely.
		p.advance()
		_ = p.collectUntilEnd(tagName)
	default:
		// Unknown tag — skip to its End.
		p.advance()
		_ = p.collectUntilEnd(tagName)
	}
}

// normalizeIndent prepares a source line for display: expands tabs to 2 spaces
// and trims trailing whitespace, but PRESERVES leading indentation (unlike
// strings.TrimSpace, which would flatten nested code). The leading indent is
// what makes Try/Catch and loop bodies readable in the docs.
func normalizeIndent(line string) string {
	// Expand tabs to 2 spaces (Xojo IDE default).
	line = strings.ReplaceAll(line, "\t", "  ")
	// Trim only trailing whitespace; keep leading indent.
	line = strings.TrimRight(line, " \t\r")
	return line
}

// dedentCommon removes the longest common leading-whitespace prefix from a
// block of source lines, so the snippet aligns to the left margin regardless
// of how deeply it was indented in the original .xojo_code file.
func dedentCommon(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	common := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue // skip blank lines when measuring
		}
		n := 0
		for n < len(l) && (l[n] == ' ') {
			n++
		}
		if common < 0 || n < common {
			common = n
		}
	}
	if common <= 0 {
		return lines
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if len(l) >= common {
			out[i] = l[common:]
		} else {
			out[i] = l
		}
	}
	return out
}

// parseMethodBody parses a Method body (Sub/Function ... End Sub/Function),
// collecting trailing #tag Note / #tag Attribute children.
func (p *tagParser) parseMethodBody(inline string) *Method {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	m := &Method{
		baseMember: baseMember{Scope: scope, IsShared: shared},
	}
	// Find the signature line (first non-blank line).
	var bodyLines []string
	sigLine := ""
	sigRaw := ""
	for p.hasNext() {
		line := p.cur()
		t := strings.TrimSpace(line)
		if t == "" {
			p.advance()
			continue
		}
		// A nested #tag appears (Note/Attribute) before any code — collect docs.
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			break
		}
		sigLine = t
		sigRaw = normalizeIndent(line)
		m.RawDecl = t
		p.advance()
		break
	}
	if sigLine != "" {
		name, isFn, params, ret, ok := parseSubFunctionLine(sigLine)
		if ok {
			m.Name = name
			m.IsFunction = isFn
			m.Params = params
			m.ReturnType = ret
		} else {
			m.Name = firstToken(sigLine)
		}
	}
	// Consume the rest of the method body until #tag EndMethod.
	var srcLines []string
	if sigRaw != "" {
		srcLines = append(srcLines, sigRaw)
	}
	for p.hasNext() {
		raw := p.cur()              // preserve indentation for source capture
		t := strings.TrimSpace(raw) // trimmed for control-flow checks
		if isEndTag(t, "Method") {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			// nested Note / Attribute
			innerTag, innerInline := parseTagName(t)
			p.advance()
			if innerTag == "Note" {
				body := p.collectUntilEnd("Note")
				applyNote(m, innerInline, body)
			} else if innerTag == "Attribute" {
				body := p.collectUntilEnd("Attribute")
				if a := parseAttributeBody(body); a != nil {
					m.Attributes = append(m.Attributes, *a)
					applyAttribute(m, *a)
				}
			} else {
				_ = p.collectUntilEnd(innerTag)
			}
			continue
		}
		bodyLines = append(bodyLines, t)
		srcLines = append(srcLines, normalizeIndent(raw))
		p.advance()
	}
	// Leading comments from the body (best-effort: take contiguous leading ' lines).
	m.DocComments = leadingComments(bodyLines)
	m.Source = strings.Join(dedentCommon(srcLines), "\n")
	m.Signature = renderMethodSignature(m)
	applyAttributes(m)
	return m
}

// parseEventHandlerBody is like parseMethodBody for #tag Event blocks.
func (p *tagParser) parseEventHandlerBody(inline string, controlName string) *EventHandler {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	e := &EventHandler{
		baseMember:  baseMember{Scope: scope, IsShared: shared},
		ControlName: controlName,
	}
	var bodyLines []string
	sigLine := ""
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if t == "" {
			p.advance()
			continue
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			break
		}
		sigLine = t
		e.RawDecl = t
		p.advance()
		break
	}
	if sigLine != "" {
		name, isFn, params, ret, ok := parseSubFunctionLine(sigLine)
		if ok {
			e.Name = name
			e.IsFunction = isFn
			e.Params = params
			e.ReturnType = ret
		} else {
			e.Name = firstToken(sigLine)
		}
	}
	var srcLines []string
	if sigLine != "" {
		srcLines = append(srcLines, sigLine)
	}
	for p.hasNext() {
		raw := p.cur()
		t := strings.TrimSpace(raw)
		if isEndTag(t, "Event") {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			if innerTag == "Note" {
				body := p.collectUntilEnd("Note")
				applyNoteEH(e, body)
			} else {
				_ = p.collectUntilEnd(innerTag)
			}
			continue
		}
		bodyLines = append(bodyLines, t)
		srcLines = append(srcLines, normalizeIndent(raw))
		p.advance()
	}
	e.DocComments = leadingComments(bodyLines)
	e.Source = strings.Join(dedentCommon(srcLines), "\n")
	return e
}

// parseEventDefBody parses #tag Hook (an event definition).
func (p *tagParser) parseEventDefBody(inline string) *EventDef {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	e := &EventDef{baseMember: baseMember{Scope: scope, IsShared: shared}}
	var srcLines []string
	for p.hasNext() {
		raw := p.cur()
		t := strings.TrimSpace(raw)
		if isEndTag(t, "Hook") {
			p.advance()
			break
		}
		if t == "" {
			p.advance()
			continue
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			_ = p.collectUntilEnd(innerTag)
			continue
		}
		srcLines = append(srcLines, normalizeIndent(raw))
		if e.RawDecl == "" {
			e.RawDecl = t
			name, isFn, params, ret, ok := parseSubFunctionLine(t)
			if ok {
				e.Name = name
				e.IsFunction = isFn
				e.Params = params
				e.ReturnType = ret
			}
		}
		p.advance()
	}
	e.Source = strings.Join(dedentCommon(srcLines), "\n")
	return e
}

// parseDelegateBody parses #tag Delegate.
func (p *tagParser) parseDelegateBody(inline string) *Delegate {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	d := &Delegate{baseMember: baseMember{Scope: scope, IsShared: shared}}
	var srcLines []string
	for p.hasNext() {
		raw := p.cur()
		t := strings.TrimSpace(raw)
		if isEndTag(t, "Delegate") {
			p.advance()
			break
		}
		if t == "" {
			p.advance()
			continue
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			_ = p.collectUntilEnd(innerTag)
			continue
		}
		srcLines = append(srcLines, normalizeIndent(raw))
		if d.RawDecl == "" {
			d.RawDecl = t
			name, isFn, params, ret, ok := parseSubFunctionLine(t)
			if ok {
				d.Name = name
				d.IsFunction = isFn
				d.Params = params
				d.ReturnType = ret
			}
		}
		p.advance()
	}
	d.Source = strings.Join(dedentCommon(srcLines), "\n")
	return d
}

// parsePropertyBody parses a #tag Property block.
func (p *tagParser) parsePropertyBody(inline string) *Property {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	pr := &Property{baseMember: baseMember{Scope: scope, IsShared: shared}}
	var bodyLines []string
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "Property") {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			if innerTag == "Note" {
				body := p.collectUntilEnd("Note")
				applyNoteProp(pr, body)
			} else {
				_ = p.collectUntilEnd(innerTag)
			}
			continue
		}
		if t == "" {
			p.advance()
			continue
		}
		if pr.RawDecl == "" {
			pr.RawDecl = t
			if name, typ, def, scopeKw, ok := parsePropertyDecl(t); ok {
				pr.Name = name
				pr.Type = typ
				pr.DefaultValue = def
				if scopeKw != "" {
					pr.Scope = scopeFromString(scopeKw)
				}
			}
		} else {
			bodyLines = append(bodyLines, t)
		}
		p.advance()
	}
	pr.DocComments = leadingComments(bodyLines)
	return pr
}

// collectAccessorSource consumes a Get/Set accessor body (already advanced
// past the opening #tag line) and returns the source as VB code: the accessor
// keyword line through the matching End line, with the #tag scaffolding
// stripped. Indentation is preserved; the common prefix is dedented. Nested
// #tag blocks (e.g. Attribute) are skipped.
func collectAccessorSource(p *tagParser, kind string) string {
	var lines []string
	endLower := strings.ToLower("#tag end" + kind)
	for p.hasNext() {
		raw := p.cur()
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(strings.ToLower(t), endLower) {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			_ = p.collectUntilEnd(innerTag)
			continue
		}
		lines = append(lines, normalizeIndent(raw))
		p.advance()
	}
	return strings.Join(dedentCommon(lines), "\n")
}

// parseComputedPropertyBody parses a #tag ComputedProperty block with
// nested #tag Getter / #tag Setter (or #tag Get / #tag Set).
func (p *tagParser) parseComputedPropertyBody(inline string) *ComputedProperty {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	cp := &ComputedProperty{baseMember: baseMember{Scope: scope, IsShared: shared}}
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "ComputedProperty") {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			switch innerTag {
			case "Getter", "Get":
				p.advance()
				cp.GetterSrc = collectAccessorSource(p, innerTag)
				cp.HasGetter = true
			case "Setter", "Set":
				p.advance()
				cp.SetterSrc = collectAccessorSource(p, innerTag)
				cp.HasSetter = true
			case "Note":
				p.advance()
				body := p.collectUntilEnd("Note")
				applyNoteCP(cp, body)
			default:
				p.advance()
				_ = p.collectUntilEnd(innerTag)
			}
			continue
		}
		if t == "" {
			p.advance()
			continue
		}
		// Header line "Name As Type".
		if cp.Type == "" {
			if name, typ, _, scopeKw, ok := parsePropertyDecl(t); ok {
				cp.Name = name
				cp.Type = typ
				if scopeKw != "" {
					cp.Scope = scopeFromString(scopeKw)
				}
			}
		}
		p.advance()
	}
	cp.IsReadOnly = cp.HasGetter && !cp.HasSetter
	return cp
}

// parseConstantBlock parses #tag Constant (single-line attrs) + optional children.
func (p *tagParser) parseConstantBlock(inline string) {
	props := parseInlineProps(inline)
	c := &Constant{baseMember: baseMember{}}
	c.Name = props["Name"]
	c.Type = props["Type"]
	c.Dynamic = strings.EqualFold(props["Dynamic"], "True")
	c.Default = unescapeConstantValue(props["Default"])
	c.Scope = scopeFromString(props["Scope"])
	if c.Scope == ScopePublic && props["Scope"] == "" {
		// Constants without Scope default to Protected in practice, but keep Public for safety.
	}
	p.container.Members = append(p.container.Members, c)
	// Consume until #tag EndConstant.
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "Constant") {
			p.advance()
			return
		}
		// skip nested Instance / Note
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			_ = p.collectUntilEnd(innerTag)
			continue
		}
		p.advance()
	}
}

// parseEnumBlock parses a standalone #tag Enum.
func (p *tagParser) parseEnumBlock(inline string) {
	props := parseInlineProps(inline)
	scope, shared := scopeFromFlags(props["Flags"])
	e := &Enum{
		baseMember: baseMember{Scope: scope, IsShared: shared},
		Type:       props["Type"],
	}
	e.Name = props["Name"]
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "Enum") {
			p.advance()
			break
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			innerTag, _ := parseTagName(t)
			p.advance()
			_ = p.collectUntilEnd(innerTag)
			continue
		}
		if t == "" {
			p.advance()
			continue
		}
		// Member line: "Name [= value]"
		em := EnumMember{}
		if eq := indexEqualsOutsideQuotes(t); eq >= 0 {
			em.Name = strings.TrimSpace(t[:eq])
			em.Value = stripQuotes(strings.TrimSpace(t[eq+1:]))
		} else {
			em.Name = t
		}
		e.Members = append(e.Members, em)
		p.advance()
	}
	p.container.Members = append(p.container.Members, e)
}

// parseWindowCodeSection parses #tag WindowCode (page/window-level code).
func (p *tagParser) parseWindowCodeSection() {
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "WindowCode") {
			p.advance()
			return
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			p.parseMemberOrSection("WindowCode")
			continue
		}
		p.advance()
	}
}

// parseEventsSection parses "#tag Events <ControlName>" — a block of event
// handlers for one control.
func (p *tagParser) parseEventsSection(controlName string) {
	for p.hasNext() {
		t := strings.TrimSpace(p.cur())
		if isEndTag(t, "Events") {
			p.advance()
			return
		}
		if strings.HasPrefix(t, "#tag ") || strings.HasPrefix(t, "#Tag ") {
			tagName, inline := parseTagName(t)
			if tagName == "Event" {
				p.advance()
				eh := p.parseEventHandlerBody(inline, controlName)
				if eh != nil {
					p.container.Members = append(p.container.Members, eh)
				}
				continue
			}
			// other tags: skip
			p.advance()
			_ = p.collectUntilEnd(tagName)
			continue
		}
		p.advance()
	}
}

// parseBeginBlock parses a "Begin <Type> <Name> ... End" designer block,
// returning a Control. Used for the UI layout on pages/menus/toolbars.
func (p *tagParser) parseBeginBlock() *Control {
	header := strings.TrimSpace(p.cur())
	// "Begin WebButton LoginButton"
	parts := strings.Fields(header)
	if len(parts) < 3 {
		p.advance()
		return nil
	}
	ctrlType := parts[1]
	ctrlName := parts[2]
	ctrl := &Control{Type: ctrlType, Name: ctrlName, Properties: map[string]string{}}
	p.advance()
	depth := 1
	for p.hasNext() && depth > 0 {
		t := strings.TrimSpace(p.cur())
		switch {
		case strings.HasPrefix(t, "Begin "):
			// Nested control — parse it as a child.
			child := p.parseBeginBlock()
			if child != nil {
				ctrl.Children = append(ctrl.Children, child)
			}
			continue // parseBeginBlock advanced past its own End
		case t == "End":
			depth--
			if depth == 0 {
				p.advance()
				return ctrl
			}
			p.advance()
		default:
			// property line
			if k, v, ok := parseKeyValueLine(t); ok && !strings.HasPrefix(k, "_") {
				ctrl.Properties[k] = v
			}
			p.advance()
		}
	}
	return ctrl
}

// collectUntilEnd consumes lines until a "#tag End<name>" line (which is also
// consumed). Returns the concatenated body text.
func (p *tagParser) collectUntilEnd(tagName string) string {
	var b strings.Builder
	endLower := strings.ToLower("#tag end" + tagName)
	for p.hasNext() {
		raw := p.cur()
		t := strings.TrimSpace(raw)
		if strings.HasPrefix(strings.ToLower(t), endLower) {
			p.advance()
			return strings.TrimRight(b.String(), "\n")
		}
		// Some tags use a single-token closer we may have to be lenient about.
		b.WriteString(t)
		b.WriteByte('\n')
		p.advance()
	}
	return strings.TrimRight(b.String(), "\n")
}

// ---- helpers shared across the parser ----

// parseTagName returns the tag keyword and the trailing inline text from a
// "#tag Foo, Bar = Baz" line. The tag name ends at the first space OR comma
// (whichever comes first), so "Method, Flags = &h0" yields name="Method".
func parseTagName(line string) (name, inline string) {
	line = strings.TrimSpace(line)
	// strip "#tag " or "#Tag "
	if strings.HasPrefix(line, "#tag ") {
		line = line[5:]
	} else if strings.HasPrefix(line, "#Tag ") {
		line = line[5:]
	} else if strings.EqualFold(line, "#tag") || strings.EqualFold(line, "#Tag") {
		return "", ""
	}
	line = strings.TrimSpace(line)
	// tag name is the first token (alphanum); it ends at the first space or comma.
	sp := strings.IndexByte(line, ' ')
	cm := strings.IndexByte(line, ',')
	cut := -1
	switch {
	case sp >= 0 && cm >= 0:
		if sp < cm {
			cut = sp
		} else {
			cut = cm
		}
	case sp >= 0:
		cut = sp
	case cm >= 0:
		cut = cm
	}
	if cut < 0 {
		return line, ""
	}
	return line[:cut], line[cut:]
}

// isEndTag reports whether the line is "#tag End<name>" for the given tag.
func isEndTag(line, tagName string) bool {
	line = strings.TrimSpace(line)
	endTag := strings.ToLower("#tag end" + tagName)
	return strings.HasPrefix(strings.ToLower(line), endTag)
}

// scopeFromFlags parses a "Flags = &h.." hex value into a Scope and shared flag.
func scopeFromFlags(flagsStr string) (Scope, bool) {
	flagsStr = strings.TrimSpace(flagsStr)
	if flagsStr == "" {
		return ScopePublic, false
	}
	flagsStr = strings.TrimPrefix(flagsStr, "&h")
	flagsStr = strings.TrimPrefix(flagsStr, "&H")
	v, err := parseHexU64(flagsStr)
	if err != nil {
		return ScopePublic, false
	}
	scope := Scope(v & 0x3)
	shared := (v & 0x20) != 0
	return scope, shared
}

func parseHexU64(s string) (uint64, error) {
	return parseHexID(s)
}

// scopeFromString maps a named scope keyword.
func scopeFromString(s string) Scope {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "private":
		return ScopePrivate
	case "protected":
		return ScopeProtected
	case "public":
		return ScopePublic
	}
	return ScopePublic
}

func firstToken(s string) string {
	s = strings.TrimSpace(s)
	if sp := strings.IndexByte(s, ' '); sp >= 0 {
		return s[:sp]
	}
	return s
}

// leadingComments returns contiguous ' comment lines from the top of bodyLines,
// stopping at the first non-comment line.
func leadingComments(bodyLines []string) []string {
	var out []string
	for _, l := range bodyLines {
		tl := strings.TrimSpace(l)
		if strings.HasPrefix(tl, "'") || strings.HasPrefix(tl, "//") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(tl, "'"), "//")))
			continue
		}
		if tl == "" {
			continue
		}
		break
	}
	return out
}

// splitInterfaces splits an "Implements A, B" tail.
func splitInterfaces(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ---- doc-extraction helpers (used by docs.go and the parser) ----

func applyNote(m *Method, inline, body string) {
	props := parseInlineProps(inline)
	if name, ok := props["Name"]; ok {
		m.NamedNotes = append(m.NamedNotes, NamedNote{Name: name, Body: body})
	} else {
		if m.Notes != "" {
			m.Notes += "\n\n"
		}
		m.Notes += body
	}
}
func applyNoteProp(p *Property, body string) {
	if p.Notes != "" {
		p.Notes += "\n\n"
	}
	p.Notes += body
}
func applyNoteCP(c *ComputedProperty, body string) {
	if c.Notes != "" {
		c.Notes += "\n\n"
	}
	c.Notes += body
}
func applyNoteEH(e *EventHandler, body string) {
	if e.Notes != "" {
		e.Notes += "\n\n"
	}
	e.Notes += body
}

// parseAttributeBody parses the body of a #tag Attribute block, which is
// "Name = Value" (Value optional).
func parseAttributeBody(body string) *Attribute {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	if eq := indexEqualsOutsideQuotes(body); eq >= 0 {
		return &Attribute{
			Name:  strings.TrimSpace(body[:eq]),
			Value: stripQuotes(strings.TrimSpace(body[eq+1:])),
		}
	}
	return &Attribute{Name: body}
}

func applyAttribute(m *Method, a Attribute) {
	if strings.EqualFold(a.Name, "Deprecated") {
		m.IsDeprecated = true
		m.DeprecMsg = a.Value
	}
}
func applyAttributes(m *Method) {
	for _, a := range m.Attributes {
		if strings.EqualFold(a.Name, "Deprecated") {
			m.IsDeprecated = true
			m.DeprecMsg = a.Value
		}
	}
}

// renderMethodSignature builds a normalized display signature for a method.
func renderMethodSignature(m *Method) string {
	var b strings.Builder
	if m.IsFunction {
		b.WriteString("Function ")
	} else {
		b.WriteString("Sub ")
	}
	b.WriteString(m.Name)
	b.WriteByte('(')
	for i, p := range m.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(renderParam(p))
	}
	b.WriteByte(')')
	if m.IsFunction && m.ReturnType != "" {
		b.WriteString(" As ")
		b.WriteString(m.ReturnType)
	}
	return b.String()
}

func renderParam(p Param) string {
	var b strings.Builder
	if p.ByRef {
		b.WriteString("ByRef ")
	}
	if p.ByVal {
		b.WriteString("ByVal ")
	}
	if p.Optional {
		b.WriteString("Optional ")
	}
	if p.ParamArray {
		b.WriteString("ParamArray ")
	}
	if p.Assigns {
		b.WriteString("Assigns ")
	}
	if p.Extends {
		b.WriteString("Extends ")
	}
	b.WriteString(p.Name)
	if p.Type != "" {
		b.WriteString(" As ")
		b.WriteString(p.Type)
	}
	if p.Default != "" {
		b.WriteString(" = ")
		b.WriteString(p.Default)
	}
	return b.String()
}

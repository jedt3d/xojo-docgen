package main

// Data model for the extracted Xojo project.
//
// The extractor reads .xojo_project (manifest) + .xojo_code/.xojo_window/etc.
// and produces this model. The renderer then walks it to emit Markdown.

// Scope of a member or class.
type Scope int

const (
	ScopePublic Scope = iota
	ScopeProtected
	ScopePrivate
)

func (s Scope) String() string {
	switch s {
	case ScopePublic:
		return "Public"
	case ScopeProtected:
		return "Protected"
	case ScopePrivate:
		return "Private"
	}
	return "Public"
}

// Project is the top-level model for one parsed Xojo project.
type Project struct {
	Name          string                // human name (from .xojo_project filename or DefaultWindow)
	Slug          string                // filesystem-safe slug, used for docs/api/<slug>/
	Type          string                // Type= value (Console, Desktop, iOS, Mobile, Web2)
	RBVersion     string                // RBProjectVersion=
	Config        map[string]string     // full key=value config block from .xojo_project
	ManifestPath  string                // absolute path to the .xojo_project
	ProjectDir    string                // directory containing the .xojo_project
	RootItems     []*Container          // top-level items in tree form
	ItemsByID     map[uint64]*Container // ID -> container (manifest lookup)
	AllContainers []*Container          // flat list of all containers (for rendering)
	Notes         string                // project-level note (if any)
}

// ManifestItem is one ItemType=Name;Path;ID;ParentID;Visible line.
type ManifestItem struct {
	ItemType string
	Name     string
	Path     string
	ID       uint64
	ParentID uint64
	Visible  bool
}

// ContainerKind classifies what a container is (maps to ItemType prefixes).
type ContainerKind int

const (
	KindUnknown ContainerKind = iota
	KindClass
	KindModule
	KindInterface
	KindWebSession
	KindPage // WebPage / MobileScreen / DesktopWindow / iOSLayout / container / dialog
	KindMenuBar
	KindToolbar
	KindFolder // pure grouping — contributes no namespace segment
	KindLibrary
	KindBuildSteps
	KindOther // MultiImage, AppIcons, ColorAsset, etc. — not documented
)

// Container is a node in the project tree: a class, module, page, menu, etc.
type Container struct {
	ItemType     string // raw ItemType prefix from manifest
	Name         string // local name
	FQN          string // fully-qualified name (namespace path)
	Kind         ContainerKind
	Scope        Scope
	Super        string   // Inherits <Super> (class) — empty if none
	Implements   []string // Implements/Aggregates interfaces
	ScopeKw      string   // raw scope keyword from body ("Public"/"Protected"/"Private")
	Members      []Member
	Children     []*Container
	Parent       *Container
	Notes        string       // concatenated unnamed #tag Note blocks
	NamedNotes   []NamedNote  // named #tag Note blocks (e.g. "Version Info")
	DocComments  []string     // leading ' comment lines at the top of the body
	Attributes   []Attribute  // Deprecated, LibraryDescription, etc.
	SourceFile   string       // relative path to the parsed file
	Controls     []*Control   // for pages/windows: the Begin/End UI tree (top-level controls)
	ManifestItem ManifestItem // the manifest line this came from
	// kindLabel cache
	kindLabel string
}

// NamedNote is a #tag Note block carrying a Name= .
type NamedNote struct {
	Name string
	Body string
}

// Attribute is a #tag Attribute entry, e.g. Deprecated = "newName".
type Attribute struct {
	Name  string
	Value string
}

// Member is any member of a container.
type Member interface {
	MemberName() string
	MemberKind() string
	MemberScope() Scope
}

// baseMember embeds common fields.
type baseMember struct {
	Name         string
	Scope        Scope
	IsShared     bool
	IsDeprecated bool
	DeprecMsg    string
	Notes        string
	NamedNotes   []NamedNote
	DocComments  []string
	Attributes   []Attribute
}

// Method is a Sub or Function.
type Method struct {
	baseMember
	IsFunction bool
	Params     []Param
	ReturnType string
	Body       string // raw body source (for potential future use)
	Signature  string // normalized signature string, e.g. "Function AddInvoice(...) As Boolean"
	RawDecl    string // the first body line, verbatim
	Source     string // full source: signature + body + End line
}

func (m *Method) MemberName() string { return m.Name }
func (m *Method) MemberKind() string { return "Method" }
func (m *Method) MemberScope() Scope { return m.Scope }

// Property is a stored property.
type Property struct {
	baseMember
	Type         string
	DefaultValue string
	RawDecl      string
}

func (p *Property) MemberName() string { return p.Name }
func (p *Property) MemberKind() string { return "Property" }
func (p *Property) MemberScope() Scope { return p.Scope }

// ComputedProperty is a property with Get/Set blocks.
type ComputedProperty struct {
	baseMember
	Type       string
	HasGetter  bool
	HasSetter  bool
	IsReadOnly bool
	RawDecl    string
	GetterSrc  string // source of the Get block (including Get/End Get)
	SetterSrc  string // source of the Set block (including Set/End Set)
}

func (c *ComputedProperty) MemberName() string { return c.Name }
func (c *ComputedProperty) MemberKind() string { return "Computed Property" }
func (c *ComputedProperty) MemberScope() Scope { return c.Scope }

// Constant is a #tag Constant.
type Constant struct {
	baseMember
	Type    string
	Dynamic bool
	Default string
}

func (c *Constant) MemberName() string { return c.Name }
func (c *Constant) MemberKind() string { return "Constant" }
func (c *Constant) MemberScope() Scope { return c.Scope }

// Enum is a standalone language Enum (#tag Enum).
type Enum struct {
	baseMember
	Type    string
	Members []EnumMember
}

func (e *Enum) MemberName() string { return e.Name }
func (e *Enum) MemberKind() string { return "Enum" }
func (e *Enum) MemberScope() Scope { return e.Scope }

// EnumMember is one value in an Enum.
type EnumMember struct {
	Name  string
	Value string
}

// Delegate is a #tag Delegate.
type Delegate struct {
	baseMember
	IsFunction bool
	Params     []Param
	ReturnType string
	RawDecl    string
	Source     string // full source including signature + body + End line
}

func (d *Delegate) MemberName() string { return d.Name }
func (d *Delegate) MemberKind() string { return "Delegate" }
func (d *Delegate) MemberScope() Scope { return d.Scope }

// EventDef is an event definition declared on a class (#tag Hook).
type EventDef struct {
	baseMember
	IsFunction bool
	Params     []Param
	ReturnType string
	RawDecl    string
	Source     string // full source including signature + body + End line
}

func (e *EventDef) MemberName() string { return e.Name }
func (e *EventDef) MemberKind() string { return "Event Definition" }
func (e *EventDef) MemberScope() Scope { return e.Scope }

// EventHandler is an implemented event handler (#tag Event).
type EventHandler struct {
	baseMember
	IsFunction  bool
	Params      []Param
	ReturnType  string
	RawDecl     string
	ControlName string // name of the control this handler belongs to (for #tag Events blocks); empty for page-level
	Source      string // full source including signature + body + End line
}

func (e *EventHandler) MemberName() string { return e.Name }
func (e *EventHandler) MemberKind() string { return "Event Handler" }
func (e *EventHandler) MemberScope() Scope { return e.Scope }

// Control is a UI control from a Begin/End block on a page/window.
type Control struct {
	Type       string // e.g. "WebButton", "DesktopLabel"
	Name       string // instance name
	Properties map[string]string
	Children   []*Control
}

// Param is one method/delegate parameter.
type Param struct {
	Name       string
	Type       string
	ByRef      bool
	ByVal      bool
	Optional   bool
	ParamArray bool
	Assigns    bool
	Extends    bool
	Default    string
	Raw        string // raw text of the param, for display
}

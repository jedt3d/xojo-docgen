# Xojo Naming Guide

Standard naming conventions used while reviewing DocGen output. These rules are
grounded in Xojo's code-management guidance and apply across Console, Desktop,
Web, iOS, and Mobile projects. Target-specific UI types retain their target's
native prefix and suffix.

> **Quick version:** title case for types, properties, and methods (`CustomerName`, `SaveCustomer`) · camel case for locals and parameters (`customerName`) · `k` prefix for constants (`kMaxUsers`) · `m` prefix for private computed-property backing fields (`mCustomerName`) · control name = purpose + type suffix (`SaveButton`, `NameField`).

---

## 1. Principles

1. **Consistency beats cleverness.** A predictable name is more valuable than a clever one. Pick the pattern and apply it everywhere.
2. **Spell things out.** Minimize abbreviations. Use `customerName`, not `custNm`. Single letters are fine only for loop counters (`i`, `j`, `c`).
3. **Encode intent, not implementation.** A name should describe *what* the thing is or does, not how it's stored or typed internally.
4. **Suffixes carry type information** for controls, views, and classes — so a reader can infer the kind from the name alone.

---

## 2. Casing

Xojo documentation defines two casings:

| Term | Rule | Example |
|------|------|---------|
| **camelCase** | First word lowercase; subsequent words capitalized | `customerName` |
| **TitleCase** (a.k.a. PascalCase) | Every word capitalized | `CustomerName` |

Keywords (`Var`, `If`, `For Each`) and data types (`Integer`, `String`) are also written in **TitleCase**.

---

## 3. Identifier rules

### Constants
- Start with lowercase **`k`**, then TitleCase.
- Examples: `kMaxUsers`, `kDefaultTimeout`, `kApiBaseUrl`.

```vb
Const kMaxUsers As Integer = 100
```

### Local variables
- **camelCase.**
- Spell out meaning; avoid abbreviations.

```vb
Var customerName As String
Var invoiceCount As Integer
```

### Arrays
- **Plural**, camelCase for locals, TitleCase for properties.

```vb
Var customers() As String          ' local
Customers() As String              ' property
```

### Properties
- **TitleCase.**

```vb
CustomerName As String
InvoiceTotal As Currency
```

### Computed properties and their backing fields
- The computed property itself is **TitleCase**: `CustomerName`.
- A `Private` backing field paired with a computed property starts with **`m`** then TitleCase: `mCustomerName`.

```vb
' Private backing field
Private mCustomerName As String

' Computed Property
ComputedProperty CustomerName As String
   Get
      Return mCustomerName
   End Get
   Set
      mCustomerName = value
   End Set
End ComputedProperty
```

### Methods
- **TitleCase.**
- Parameters are **camelCase**.

```vb
Sub SaveCustomer(customerName As String, invoiceId As Integer)
End Sub

Function FindInvoice(invoiceId As Integer) As Invoice
End Function
```

### Events and event handlers
- Use the modern Xojo event names: `Opening`, `Closing`, `Pressed`, `Shown`, `TextChanged` — **not** the legacy `Open`, `Close`, `Action`.
- Event handlers on a control keep the event name: a `WebButton.Pressed` handler is `SaveButton_Pressed` (Xojo generates `<ControlName>_<EventName>`).

---

## 4. Control naming

**Pattern:** `[Purpose][TypeSuffix]` — TitleCase, with a suffix that identifies the control type.

Infer the purpose from the control's most meaningful text, in priority order: **Caption → Hint → Tooltip → InitialValue**. Every control name must be **unique within its parent** (page, container, or dialog).

### Control suffix table

The suffixes below follow the Xojo documentation where it defines one and add
project-specific semantic suffixes for controls it does not list:

| Control | Suffix | Example |
|---------|--------|---------|
| `WebButton` | `Button` | `SaveButton` |
| `WebLink` | `Link` | `SignUpLink` |
| `WebListBox` | `List` | `InvoiceList` |
| `WebSegmentedButton` | `Selector` | `TaskSelector` |
| `WebCheckBox` | `Check` | `TaxableCheck` |
| `WebPopupMenu` | `Popup` | `StatusPopup` |
| `WebRadioGroup` | `Radio` | `SourceRadio` |
| `WebTextField`, `WebSearchField` | `Field` | `UsernameField` |
| `WebTextArea` | `Area` | `DescriptionArea` |
| `WebCanvas` | `Canvas` | `ChartCanvas` |
| `WebLabel` | `Label` | `NameLabel` |
| `WebPagePanel` | `Panel` | `MainPanel` |
| `WebTabPanel` | `Tab` | `SettingsTab` |
| `WebProgressBar`, `WebProgressWheel` | `Progress` | `DownloadProgress` |
| `WebHTMLViewer` | `Viewer` | `DocViewer` |
| `WebImageViewer` | `Image` | `ProfileImage` |
| `WebSlider` | `Slider` | `VolumeSlider` |
| `WebSwitch` | `Switch` | `NotificationsSwitch` |
| `WebRectangle` | `Rectangle` | `DividerRectangle` |
| `WebToolbar` | `Toolbar` | `MainToolbar` |
| `WebFlexLayoutManager` | `Layout` | `LoginFormLayout` |

For a custom subclass or compound `WebContainer`, name each placed instance by
its semantic base control, not by its vendor prefix or implementation class. For
example, an `XjWebTextField` used for a username is `UsernameField`, not
`UsernameXjWebTextField` or `UsernameContainer`. This keeps callers independent
of the implementation while the project item itself retains its class name.

> **Two conventions in circulation.** The official Xojo suffixes above use *shortened* type words (`WebListBox` → `…List`, `WebTextField` → `…Field`). The bundled Xojo skills document a simpler "drop the `Web` prefix" rule (`WebListBox` → `InvoiceListBox`, keeping the full type name). **This project uses the official shortened suffixes** in the table above. Pick the one your team prefers and apply it consistently — mixing the two is worse than either one alone.

---

## 5. Type naming — classes, modules, interfaces

### Plain classes
- **TitleCase.**
- Subclasses of a built-in class keep the built-in name as the suffix: `CustomerListBox`, `InvoiceWebView`.

### Interfaces
- **TitleCase with the `Interface` suffix** (per official Xojo docs): `PaymentInterface`, `RepositoryInterface`.

> An alternative convention uses an `I` prefix (`IPayment`, `IRepository`).
> This guide follows Xojo's `Interface` suffix. Whichever convention a project
> selects, apply it consistently rather than mixing both.

### Modules
- **TitleCase.**
- A descriptive name is fine without a `Module` suffix, but adding one is acceptable for clarity: `UtilityModule`, `Networking`, `Constants`.

### Class delegates and custom exceptions
- Custom exceptions are TitleCase classes, typically with an `Exception` suffix: `InvoiceNotFoundException`, `PdfRenderException`.

---

## 6. UI target type naming

Use the semantic purpose followed by the native Xojo type suffix:
`[Purpose][TypeSuffix]`.

| Target/type | Suffix | Example |
|---|---|---|
| Desktop window (`DesktopWindow`) | `Window` | `DashboardWindow`, `SettingsWindow` |
| Web page (`WebPage`) | `Page` | `DashboardPage`, `SettingsPage` |
| Mobile/iOS screen | `Screen` | `CustomerScreen`, `InvoiceScreen` |
| Container | `Container` | `LoginContainer`, `InvoiceLineContainer` |
| Dialog | `Dialog` | `ConfirmDialog`, `ExportDialog` |
| Toolbar | `Toolbar` | `CustomerDetailsToolbar` |
| Custom SDK control | domain suffix | `TextCounter`, `BarcodeScanner` |
| Style (`WebStyle`) | `Style` | `PrimaryButtonStyle`, `ErrorLabelStyle` |

**App** is special and normally retains that name. Web projects also use
**Session** for the `WebSession` subclass. These names are registered in the
`.xojo_project` manifest; renaming them requires coordinated manifest and item
changes.

---

## 7. Database naming

- **Table names:** plural, lowercase or snake_case as the schema dictates — follow the existing database convention, not Xojo's.
- **Xojo-side record/model classes:** TitleCase, singular: `Customer`, `Invoice`, `InvoiceLine`.
- **Database connections** should have a clear ownership and lifetime. In Web
  projects they normally live on `Session` so each user has an isolated
  connection. Name the property clearly: `DB As Database` or
  `PostgresDB As PostgreSQLDatabase`.

---

## 8. File naming

Xojo generates one external file per project item. The file name follows the item name; keep them in sync.

| Extension | Named after | Example |
|-----------|-------------|---------|
| `.xojo_project` | The project | `MyApplication.xojo_project` |
| `.xojo_code` | The class/module/interface | `App.xojo_code`, `Session.xojo_code` |
| `.xojo_window` / `.xojo_code` | A UI item, depending on export format | `DashboardWindow.xojo_window`, `DashboardPage.xojo_code` |
| `.xojo_menu` | The menu bar | `MainMenuBar.xojo_menu` |
| `.xojo_toolbar` | The toolbar | `MainToolbar.xojo_toolbar` |

**Never read or modify** the binary files: `.xojo_resources`, `.xojo_uistate`.

---

## 9. Formatting (in the code)

- **Keywords in TitleCase:** `Var`, `If`, `For`, `Each`, `Try`, `Catch`, `Return`.
- **Data types in TitleCase:** `Integer`, `String`, `Dictionary`, `DateTime`.
- **Spaces between every item** in argument and parameter lists: `SaveCustomer(name, invoiceId)`, not `SaveCustomer(name,invoiceId)`.
- **Modern syntax only:** `Var` (not `Dim`), `Opening`/`Closing`/`Pressed` (not `Open`/`Close`/`Action`), `String.Middle()`/`String.IndexOf()` (not `Mid()`/`InStr()`).

---

## 10. Quick reference

| Thing | Case / Prefix | Example |
|-------|---------------|---------|
| Constant | `k` + TitleCase | `kMaxUsers` |
| Local variable | camelCase | `customerName` |
| Method parameter | camelCase | `customerName` |
| Array (local) | camelCase, plural | `customers()` |
| Property | TitleCase | `CustomerName` |
| Computed-property backing field | `m` + TitleCase | `mCustomerName` |
| Method | TitleCase | `SaveCustomer` |
| Control | Purpose + suffix | `SaveButton`, `NameField` |
| Class | TitleCase | `Invoice`, `CustomerListBox` |
| Interface | TitleCase + `Interface` | `PaymentInterface` |
| Module | TitleCase | `UtilityModule` |
| Page | Purpose + `Page` | `DashboardPage` |
| Container | Purpose + `Container` | `LoginContainer` |
| Dialog | Purpose + `Dialog` | `ConfirmDialog` |
| Custom exception | TitleCase + `Exception` | `InvoiceNotFoundException` |

---

## 11. Decision log — where this guide resolves ambiguity

The bundled Xojo skills and the official docs disagree in two spots. This project resolves them as follows:

1. **Control suffixes.** Official docs use shortened suffixes (`WebListBox` → `…List`, `WebTextField` → `…Field`); the skills use "drop the `Web` prefix" (`…ListBox`, `…TextField`). **This project uses the official shortened suffixes** (§4).
2. **Interface naming.** Official docs use the `Interface` suffix (`PaymentInterface`); the skills use an `I` prefix (`IPayment`). **This project uses the `Interface` suffix** (§5).

Both decisions favor the official Xojo documentation as the source of truth. If the team later prefers the alternatives, change the guide once and migrate the whole project — do not mix conventions.

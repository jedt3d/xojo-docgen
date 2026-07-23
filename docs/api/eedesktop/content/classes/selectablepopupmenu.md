---
title: SelectablePopupMenu
---

# SelectablePopupMenu

`Class` `Protected`

**Inherits:** [DesktopPopupMenu](https://documentation.xojo.com/api/user_interface/desktop/desktoppopupmenu.html)  

## Event Definitions

### 

`Event Definition`

<pre><code>Event Open()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Event Open()</code></pre>

</details>

## Event Handlers

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening()
      Open

    End Sub</code></pre>

</details>

## Methods

### SelectValue

`Public`

Search for the specified value in the list and select it.
If no value is found then nothing is selected.

<pre><code class="language-xojo">Sub SelectValue(value As String)
  // Search for the specified value in the list and select it.
  // If no value is found then nothing is selected.
  For i As Integer = 0 To Self.RowCount - 1
    If Self.RowValueAt(i) = value Then
      Self.SelectedRowIndex = i
      Return
    End If
  Next

  Self.SelectedRowIndex = -1

  Return
End Sub</code></pre>


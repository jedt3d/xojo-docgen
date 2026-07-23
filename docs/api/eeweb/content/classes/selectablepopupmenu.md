---
title: SelectablePopupMenu
---

# SelectablePopupMenu

`Class` `Protected`

**Inherits:** [WebPopupMenu](https://documentation.xojo.com/api/user_interface/web/webpopupmenu.html)  

## Methods

### SelectValue

`Public`

Search for the specified value in the list and select it.
If no value is found then nothing is selected.

<pre><code class="language-xojo">Sub SelectValue(value As String)
  // Search for the specified value in the list and select it.
  // If no value is found then nothing is selected.

  Self.SelectedRowIndex = -1
  For i As Integer = 0 To Self.LastRowIndex
    If Self.RowValueAt(i) = value Then SelectedRowIndex = i
  Next
End Sub</code></pre>


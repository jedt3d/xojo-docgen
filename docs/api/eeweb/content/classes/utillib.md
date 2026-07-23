---
title: UtilLib
---

# UtilLib

`Module` `Protected`

## Methods

### Filter

`Public`

<pre><code class="language-xojo">Function Filter(source As String, allowable As String) As String
  Var i As Integer
  Var result As String

  For i = 0 To source.Length - 1
    If allowable.IndexOf(source.Middle(i, 1)) &gt;= 0 Then
      result = result + source.Middle(i, 1)
    End If
  Next

  Return result
End Function</code></pre>


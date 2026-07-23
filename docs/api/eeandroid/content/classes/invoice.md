---
title: Invoice
---

# Invoice

`Class` `Protected`

## Methods

### GetMonthlyInvoiceTotalsByYear

`Public`

<pre><code>Sub GetMonthlyInvoiceTotalsByYear(year As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Shared Sub GetMonthlyInvoiceTotalsByYear(year As String)
  Amounts.RemoveAll
  For i As Integer = 0 To 11
    Amounts.Add(0)
  Next

  // For the specified year, group the invoices by month and then sum the amounts for the month
  Var sql As String = "SELECT SubStr(InvoiceDate, 6, 2), Sum(InvoiceAmount) FROM invoices WHERE substr(invoicedate, 1, 4)  = ? GROUP BY SubStr(InvoiceDate, 6, 2) ORDER BY SubStr(InvoiceDate, 6, 2)"

  Var yearVar As Variant = year
  Var rs As RowSet = App().EEDB.SelectSQL(sql, yearVar)

  If rs &lt;&gt; Nil Then
    Var month As Integer
    Var monthStr As String

    While Not rs.AfterLastRow
      // Put the amount for the month into the array
      monthStr = rs.ColumnAt(0).StringValue
      month = Integer.FromString(monthStr)

      Amounts(month - 1) = rs.ColumnAt(1).IntegerValue

      rs.MoveToNextRow
    Wend
    rs.Close
  End If

End Sub</code></pre>

</details>

## Properties

### 

`Public`

<pre><code>As</code></pre>


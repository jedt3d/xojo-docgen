---
title: SalesChart
---

# SalesChart

`Class` `Protected`

**Inherits:** [mobileChart](https://documentation.xojo.com/api/user_interface/mobile/mobilechart.html)  

## Methods

### GetMonthlyInvoiceTotalsByYear

`Protected, Shared`

For the specified year, group the invoices by month and then sum the amounts for the month

<pre><code class="language-xojo">Private Function GetMonthlyInvoiceTotalsByYear(year As String) As Double()
  // For the specified year, group the invoices by month and then sum the amounts for the month
  Var sql As String = "SELECT SubStr(InvoiceDate, 6, 2), Sum(InvoiceAmount) FROM invoices WHERE substr(invoicedate, 1, 4)  = ? GROUP BY SubStr(InvoiceDate, 6, 2) ORDER BY SubStr(InvoiceDate, 6, 2)"

  Try
    Var rs As RowSet = App.EEDB.SelectSQL(sql, year)

    Var amounts(11) As Double

    If rs &lt;&gt; Nil Then
      While Not rs.AfterLastRow
        // Put the amount for the month into the array
        Var month As Integer
        month = Integer.FromString(rs.ColumnAt(0).StringValue)
        amounts(month - 1) = rs.ColumnAt(1).CurrencyValue

        rs.MoveToNextRow
      Wend
      rs.Close
    End If

    Return amounts
  Catch e As DatabaseException
    Return Nil
  End Try
End Function</code></pre>

### LoadData

`Public`

<pre><code class="language-xojo">Sub LoadData(Year as String)
  Me.RemoveAllDatasets

  Var amounts() As Double
  amounts = GetMonthlyInvoiceTotalsByYear(year)

  Var ds As New ChartLinearDataset("Sales", Color.Blue, True, amounts)
  me.AddDataset ds
End Sub</code></pre>


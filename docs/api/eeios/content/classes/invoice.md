---
title: Invoice
---

# Invoice

`Class` `Protected`

## Methods

### GetYears

`Public`

For the specified year, group the invoices by month and then sum the amounts for the month

<pre><code class="language-xojo">Shared Function GetYears() As String()
  // For the specified year, group the invoices by month and then sum the amounts for the month
  Var sql As String = "SELECT DISTINCT SubStr(InvoiceDate, 1, 4) FROM Invoices ORDER BY InvoiceDate DESC"

  Try
    Var rs As RowSet = App.EEDB.SelectSQL(sql)

    Var years() As String

    If rs &lt;&gt; Nil Then
      While Not rs.AfterLastRow
        // Put the year into the array
        years.Add(rs.ColumnAt(0).StringValue)
        rs.MoveToNextRow
      Wend
      rs.Close
    End If

    Return years
  Catch e As DatabaseException
    Return Nil
  End Try
End Function</code></pre>

## Properties

### Customer

`Public`

<pre><code>Customer As <a href="../customer/">Customer</a></code></pre>

### InvoiceAmount

`Public`

<pre><code>InvoiceAmount As <a href="https://documentation.xojo.com/api/data_types/currency.html">Currency</a></code></pre>

### InvoiceDate

`Public`

<pre><code>InvoiceDate As <a href="https://documentation.xojo.com/api/data_types/datetime.html">DateTime</a></code></pre>

### InvoiceNumber

`Public`

<pre><code>InvoiceNumber As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>


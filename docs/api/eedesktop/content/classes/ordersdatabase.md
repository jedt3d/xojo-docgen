---
title: OrdersDatabase
---

# OrdersDatabase

`Class` `Protected`

**Inherits:** [SQLiteDatabase](https://documentation.xojo.com/api/databases/sqlitedatabase.html)  

## Methods

### AddInvoice

`Public`

<pre><code class="language-xojo">Function AddInvoice(invoiceNum As String, invoiceDate As String, invoiceAmount As Currency, customerID As String) As Boolean
  Try
    Var invoice As New DatabaseRow
    invoice.Column("InvoiceNo").StringValue = invoiceNum
    invoice.Column("InvoiceDate").StringValue = invoiceDate
    invoice.Column("InvoiceAmount").CurrencyValue = invoiceAmount
    invoice.Column("CustomerID").StringValue = customerID

    Self.AddRow("Invoices", invoice)
  Catch e As DatabaseException
    Return False
  End Try

  Return True

End Function</code></pre>

### AddInvoiceItem

`Public`

<pre><code class="language-xojo">Function AddInvoiceItem(code As String, quantity As Integer, invoiceNum As String) As Boolean
  Try
    Var invoiceRecord As New DatabaseRow

    invoiceRecord.Column("InvoiceNo").StringValue = invoiceNum
    invoiceRecord.Column("ProductCode").StringValue = code
    invoiceRecord.Column("Quantity").IntegerValue = quantity

    Self.AddRow("InvoiceItems", invoiceRecord)
  Catch e As DatabaseException
    Return False
  End Try

  Return True
End Function</code></pre>

### CancelTransaction

`Public`

<pre><code class="language-xojo">Sub CancelTransaction()
  Self.RollbackTransaction
End Sub</code></pre>

### DeleteInvoiceItems

`Public`

<pre><code class="language-xojo">Function DeleteInvoiceItems(invoiceNum As String) As Boolean
  Try
    Var sql As String = "DELETE FROM InvoiceItems WHERE InvoiceNo=?"

    Self.ExecuteSQL(sql, invoiceNum)
  Catch e As DatabaseException
    Return False
  End Try

  Return True
End Function</code></pre>

### EndTransaction

`Public`

<pre><code class="language-xojo">Sub EndTransaction()
  Self.CommitTransaction
End Sub</code></pre>

### FindCustomersByID

`Public`

<pre><code class="language-xojo">Function FindCustomersByID(ID As String) As RowSet
  Var sql As String = "SELECT * FROM Customers WHERE ID=? ORDER BY lastname, firstname"
  Var rs As RowSet = Self.SelectSQL(sql, ID)

  Return rs
End Function</code></pre>

### FindCustomersByName

`Public`

<pre><code class="language-xojo">Function FindCustomersByName(Optional searchName As String) As RowSet
  Var sql As String = "SELECT * FROM Customers WHERE lastname LIKE ? OR firstname LIKE ? ORDER BY lastname, firstname"
  Var rs As RowSet = Self.SelectSQL(sql, searchName + "%", searchName + "%")

  Return rs

End Function</code></pre>

### GetInvoiceByNumber

`Public`

<pre><code class="language-xojo">Function GetInvoiceByNumber(invoiceNum As String) As RowSet
  Var sql As String = "SELECT * FROM Invoices WHERE InvoiceNo=?"
  Var rs As RowSet = Self.SelectSQL(sql, invoiceNum)

  Return rs
End Function</code></pre>

### GetInvoiceItemsForInvoice

`Public`

<pre><code class="language-xojo">Function GetInvoiceItemsForInvoice(invoiceNum As String) As RowSet
  Var sql As String = "SELECT * FROM InvoiceItems INNER JOIN Products ON Products.Code = InvoiceItems.ProductCode WHERE InvoiceNo=?"
  Var rs As RowSet = Self.SelectSQL(sql, invoiceNum)

  Return rs
End Function</code></pre>

### GetInvoicesForCustomer

`Public`

Update the list of invoices to show invoices from the selected customer

<pre><code class="language-xojo">Function GetInvoicesForCustomer(CustomerID As String) As RowSet
  // Update the list of invoices to show invoices from the selected customer
  Var rs As RowSet = SelectSQL("SELECT * FROM Invoices WHERE CustomerID=?", CustomerID)

  Return rs

End Function</code></pre>

### GetInvoiceYears

`Public`

Determine how many unique years there are in the invoices table

<pre><code class="language-xojo">Function GetInvoiceYears() As String()
  // Determine how many unique years there are in the invoices table
  Var sql As String = "SELECT DISTINCT substr(invoicedate, 1, 4) FROM invoices ORDER BY invoicedate DESC"
  Var rs As RowSet = Self.SelectSQL(sql)

  Var years() As String

  If rs &lt;&gt; Nil Then
    For Each year As DatabaseRow In rs
      years.Add(year.ColumnAt(0).StringValue)
    Next
    rs.Close
  End If

  Return years
End Function</code></pre>

### GetMonthlyInvoiceTotalsByYear

`Public`

For the specified year, group the invoices by month and then sum the amounts for the month

<pre><code class="language-xojo">Function GetMonthlyInvoiceTotalsByYear(year As String) As Double()
  // For the specified year, group the invoices by month and then sum the amounts for the month
  Var sql As String = "SELECT substr(invoicedate, 6, 2), sum(invoiceamount) FROM invoices WHERE substr(invoicedate, 1, 4)  = ? GROUP BY substr(invoicedate, 6, 2)  ORDER BY substr(invoicedate, 6, 2)"
  Var rs As RowSet = Self.SelectSQL(sql, year)

  Var amounts(11) As Double

  If rs &lt;&gt; Nil Then
    For Each total As DatabaseRow In rs
      amounts(total.ColumnAt(0).IntegerValue - 1) = total.ColumnAt(1).IntegerValue
    Next
    rs.Close
  End If

  Return amounts
End Function</code></pre>

### GetNextInvoiceNumber

`Public`

<pre><code class="language-xojo">Function GetNextInvoiceNumber() As Integer
  Var rs As RowSet
  rs = Self.SelectSQL("SELECT Max(invoiceno) FROM Invoices")

  If rs &lt;&gt; Nil And Not rs.AfterLastRow Then
    Return rs.ColumnAt(0).IntegerValue + 1
  Else
    Return 1001
  End If
End Function</code></pre>

### GetProductByCode

`Public`

<pre><code class="language-xojo">Function GetProductByCode(code As String) As RowSet
  Var sql As String = "SELECT * FROM Products WHERE Code = ?"
  Var rs As RowSet = Self.SelectSQL(sql, code)

  Return rs
End Function</code></pre>

### GetProducts

`Public`

<pre><code class="language-xojo">Function GetProducts() As RowSet
  Var sql As String = "SELECT * FROM Products"

  Var rs As RowSet
  rs = Self.SelectSQL(sql)

  Return rs
End Function</code></pre>

### SetupNewDatabase

`Public`

The database file is copied to the App folder using a Build Automation step.

<pre><code class="language-xojo">Shared Function SetupNewDatabase() As OrdersDatabase
  // The database file is copied to the App folder using a Build Automation step.

  Var msg As String

  // Get to the database from Resources and copy
  // to ApplicationSupport in EEData folder.
  Var sourceDB As New FolderItem

  Var eeDB As FolderItem = SpecialFolder.Resource("EddiesElectronics.sqlite")
  If eeDB &lt;&gt; Nil And eeDB.Exists Then
    // Copy to AppData
    Var eeData As FolderItem = SpecialFolder.ApplicationData.Child("EEData")
    If Not eeData.Exists Then
      eeData.CreateFolder
    End If

    Var eeDestDB As FolderItem = eeData.Child("EddiesElectronics.sqlite")
    If eeDestDB.Exists Then eeDestDB.Remove

    eeDB.CopyTo(eeDestDB)
    sourceDB = New FolderItem(eeDestDB.NativePath, FolderItem.PathModes.Native)
  End If

  If sourceDB = Nil Or sourceDB.Exists = False Then
    msg = "Could not find EddiesElectronics.sqlite."
    #If TargetDesktop Then
      MessageBox(msg)
    #ElseIf TargetWeb Then
      App.AppendToDBErrorLog(msg)
    #Endif

    Return Nil
  End If

  // Create a blank, in-memory only database so the user can make changes without
  // affecting the on disk database
  Var orders As New OrdersDatabase
  orders.DatabaseFile = sourceDB
  If Not orders.Connect Then
    msg = "Could not connect to " + sourceDB.Name
    #If TargetDesktop Then
      MessageBox(msg)
    #ElseIf TargetWeb Then
      App.AppendToDBErrorLog(msg)
    #Endif

    Return Nil
  End If

  Return orders

End Function</code></pre>

### UpdateInvoice

`Public`

<pre><code class="language-xojo">Function UpdateInvoice(invoiceNum As String, invoiceDate As String, invoiceTotal As Currency) As Boolean
  Try
    Var sql As String = "SELECT InvoiceNo, CustomerID, InvoiceDate, InvoiceAmount FROM Invoices WHERE InvoiceNo=?"
    Var rs As RowSet = Self.SelectSQL(sql, invoiceNum)

    If rs &lt;&gt; Nil Then
      rs.EditRow
      rs.Column("InvoiceDate").StringValue = invoiceDate
      rs.Column("InvoiceAmount").CurrencyValue = invoiceTotal

      rs.SaveRow

      rs.Close
    End If
  Catch e As DatabaseException
    Return False
  End Try

  Return True
End Function</code></pre>

### UpdateInvoiceDates

`Public`

Update the invoice dates to a range in the last 4 years
This method gets a unique list of years in descending order.

<pre><code class="language-xojo">Sub UpdateInvoiceDates()
  'Update the invoice dates to a range in the last 4 years
  'This method gets a unique list of years in descending order.

  Try
    'Get a unique list of years in descending order.
    Var rs As RowSet = Self.SelectSQL("SELECT DISTINCT Substr(InvoiceDate, 1, 4) FROM Invoices ORDER BY InvoiceDate DESC")
    'Get the most recent year from that list
    Var mostRecentYear As Integer = rs.ColumnAt(0).IntegerValue
    Var today As DateTime = DateTime.Now
    'Determine the number of years between the current year and the most recent year in the database
    Var delta As Integer = today.Year - mostRecentYear
    'Loop through each year in the list of unique years
    For Each row As DatabaseRow In rs
      'Get the year from the list of unique years
      Var year As Integer = rs.ColumnAt(0).IntegerValue
      'Create a new year by adding the calculated delta to that year
      Var newYear As Integer = year + delta
      'Update all invoices by replacing the year with the new year
      Var sql As String
      sql = "UPDATE Invoices SET InvoiceDate = REPLACE(InvoiceDate, '" + year.ToString + "-', '" + newYear.ToString + "-')"
      Self.ExecuteSQL(sql)
    Next

  Catch error As DatabaseException
    System.Beep
    MessageBox("A database exception occurred in the UpdateInvoiceDates method.")
  End Try
End Sub</code></pre>


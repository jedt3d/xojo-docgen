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
  Var invoice As New DatabaseRow
  invoice.Column("InvoiceNo") = invoiceNum
  invoice.Column("InvoiceDate") = invoiceDate
  invoice.Column("InvoiceAmount") = invoiceAmount
  invoice.Column("CustomerID") = customerID

  Try
    Self.AddRow("Invoices", invoice)
  Catch e As DatabaseException
    Return False
  End Try

  Return True

End Function</code></pre>

### AddInvoiceItem

`Public`

<pre><code class="language-xojo">Function AddInvoiceItem(code As String, quantity As Integer, invoiceNum As String) As Boolean
  Var invoiceRecord As New DatabaseRow

  invoiceRecord.Column("InvoiceNo") = invoiceNum
  invoiceRecord.Column("ProductCode") = code
  invoiceRecord.Column("Quantity") = quantity

  Try
    Self.AddRow("InvoiceItems", invoiceRecord)
  Catch e As DatabaseException
    Return False
  End Try

  Return True
End Function</code></pre>

### BeginTransaction

`Public`

<pre><code class="language-xojo">Sub BeginTransaction()
  Self.ExecuteSQL("BEGIN TRANSACTION")
End Sub</code></pre>

### CancelTransaction

`Public`

<pre><code class="language-xojo">Sub CancelTransaction()
  Self.RollbackTransaction
End Sub</code></pre>

### CleanDBFolder

`Public`

This method is run at startup to clean out any session databases
left over from the last run

<pre><code class="language-xojo">Shared Sub CleanDBFolder()
  // This method is run at startup to clean out any session databases
  // left over from the last run
  Var f As FolderItem = GetDBFolder
  For Each file As FolderItem In f.Children
    file.Remove
  Next
End Sub</code></pre>

### DeleteInvoiceItems

`Public`

<pre><code class="language-xojo">Function DeleteInvoiceItems(invoiceNum As String) As Boolean
  Var sql As String
  sql = "DELETE FROM InvoiceItems WHERE InvoiceNo=" + invoiceNum

  Try
    Self.ExecuteSQL(sql)
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
  Var rs As Rowset = Self.SelectSQL("SELECT * FROM Customers WHERE ID=? ORDER BY lastname, firstname", ID)

  Return rs
End Function</code></pre>

### FindCustomersByName

`Public`

<pre><code class="language-xojo">Function FindCustomersByName(Optional searchName As String) As Rowset
  Var rs As Rowset = Self.SelectSQL("SELECT * FROM Customers WHERE lastname LIKE ? OR firstname LIKE ? ORDER BY lastname, firstname", SearchName+"%", SearchName+"%")

  Return rs

End Function</code></pre>

### Private

`Protected, Shared`

<pre><code class="language-xojo">Private Shared Function GetDBFolder() As FolderItem
  Var dbFolder As New FolderItem("Databases", FolderItem.PathModes.Native)
  If Not dbFolder.Exists Then
    dbFolder.CreateFolder

    If Not dbFolder.Exists Then
      App.AppendToDBErrorLog("DBFolder count not be created: " + dbFolder.NativePath)
    End If
  End If

  dbFolder.Permissions = &amp;o777 // Read/write permissions

  Return dbFolder
End Function</code></pre>

### GetInvoiceByNumber

`Public`

<pre><code class="language-xojo">Function GetInvoiceByNumber(invoiceNum As String) As RowSet
  Var rs As RowSet = Self.SelectSQL("SELECT * FROM Invoices WHERE InvoiceNo=?", invoiceNum)

  Return rs
End Function</code></pre>

### GetInvoiceItemsForInvoice

`Public`

<pre><code class="language-xojo">Function GetInvoiceItemsForInvoice(invoiceNum As String) As RowSet
  Var rs As Rowset = Self.SelectSQL("SELECT * FROM InvoiceItems INNER JOIN Products ON Products.Code = InvoiceItems.ProductCode WHERE InvoiceNo=?", invoiceNum)

  Return rs
End Function</code></pre>

### GetInvoicesForCustomer

`Public`

Update the list of invoices to show invoices from the selected customer

<pre><code class="language-xojo">Function GetInvoicesForCustomer(CustomerID As String) As RowSet
  //Update the list of invoices to show invoices from the selected customer
  Var rs As RowSet = Self.SelectSQL("SELECT * FROM Invoices WHERE CustomerID=? ORDER BY InvoiceNo", CustomerID)

  Return rs

End Function</code></pre>

### GetInvoiceYears

`Public`

Determine how many unique years there are in the invoices table

<pre><code class="language-xojo">Function GetInvoiceYears() As String()
  // Determine how many unique years there are in the invoices table
  Var rs As RowSet = Self.SelectSQL("SELECT DISTINCT substr(invoicedate, 1, 4) FROM invoices ORDER BY invoicedate DESC")

  Var years() As String

  If rs &lt;&gt; Nil Then
    For Each row As DatabaseRow In rs
      years.Add(rs.ColumnAt(0).StringValue)
    Next
    rs.Close
  End If

  Return years
End Function</code></pre>

### GetMonthlyInvoiceTotalsByYear

`Public`

For the specified year, group the invoices by month and then sum the amounts for the month

<pre><code class="language-xojo">Function GetMonthlyInvoiceTotalsByYear(year As String) As Integer()
  // For the specified year, group the invoices by month and then sum the amounts for the month
  Var rs As RowSet = Self.SelectSQL("SELECT substr(invoicedate, 6, 2), sum(invoiceamount) FROM invoices WHERE substr(invoicedate, 1, 4)  = ? GROUP BY substr(invoicedate, 6, 2)  ORDER BY substr(invoicedate, 6, 2)", year)

  Var amounts(11) As Integer

  If rs &lt;&gt; Nil Then
    For Each row As DatabaseRow In rs
      amounts(rs.ColumnAt(0).IntegerValue - 1) = rs.ColumnAt(1).IntegerValue
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
    Return rs.ColumnAt(1).IntegerValue + 1
  Else
    Return 1001
  End If
End Function</code></pre>

### GetProductByCode

`Public`

<pre><code class="language-xojo">Function GetProductByCode(code As String) As Rowset
  Var rs As Rowset = Self.SelectSQL("SELECT * FROM Products WHERE Code = ?", code)

  Return rs
End Function</code></pre>

### GetProducts

`Public`

<pre><code class="language-xojo">Function GetProducts() As Rowset
  Var sql As String = "SELECT * FROM Products"

  Var rs As RowSet
  rs = Self.SelectSQL(sql)

  Return rs
End Function</code></pre>

### SetupNewDatabase

`Public`

Creates a copy of the database for each user (Session) that
connects.
If you wanted to alter this example project to work with a database server and have all users share
one database, you would first change this routine to return an instance of your database server connection
instead of doing what it's doing now. Then you would replace any instances of SQLiteDatabase
In this project with the database you are using.
Alternatively, you could change this code to simply return the source database rather than make an
on-disk copy. If you are going to do that, make sure to set the SQLiteDatabase.WriteAheadLogging property
to true.

<pre><code class="language-xojo">Shared Function SetupNewDatabase() As OrdersDatabase
  // Creates a copy of the database for each user (Session) that
  // connects.

  // If you wanted to alter this example project to work with a database server and have all users share
  // one database, you would first change this routine to return an instance of your database server connection
  // instead of doing what it's doing now. Then you would replace any instances of SQLiteDatabase
  // In this project with the database you are using.

  // Alternatively, you could change this code to simply return the source database rather than make an
  // on-disk copy. If you are going to do that, make sure to set the SQLiteDatabase.WriteAheadLogging property
  // to true.

  Var msg As String

  // Make sure we can get to the database on disk
  Var source As New FolderItem("EddiesElectronics.sqlite", FolderItem.PathModes.Native)

  If source = Nil Or source.Exists = False Then
    msg = "EddiesElectronics.sqlite could not be found."
    #If TargetDesktop Then
      MessageBox(msg)
    #ElseIf TargetWeb Then
      App.AppendToDBErrorLog(msg)
    #EndIf

    Return Nil
  Else
    source.Permissions = &amp;o777
    App.AppendToDBErrorLog("Found DB: " + source.NativePath)
  End If

  // Clone the database
  Var dbFile As FolderItem = GetDBFolder.Child(Format(System.Microseconds, "0") + ".sqlite")


  Try
    source.CopyTo(dbFile)
  Catch e As IOException
    msg = "Could not clone database for this user: " + e.Message
    #If TargetDesktop Then
      MessageBox(msg)
    #ElseIf TargetWeb Then
      App.AppendToDBErrorLog(msg)
      App.AppendToDBErrorLog("File: " + dbFile.NativePath)
    #EndIf
    Return Nil
  End Try

  If Not dbFile.Exists Then
    App.AppendToDBErrorLog("DBFile does not exist: " + dbFile.NativePath)
  Else
    #If TargetXojoCloud Then
      // Workaround as setting the permissions property on a copied
      // file does not yet work.
      Var sh As New Shell
      sh.Execute("chmod 666 " + dbFile.NativePath) // read/write permissions

      App.AppendToDBErrorLog("DBFile Permissions: " + Oct(dbFile.Permissions))
    #EndIf
  End If

  Var orders As New OrdersDatabase
  orders.DatabaseFile = dbFile

  Try
    orders.Connect
  Catch e As DatabaseException
    msg = "Could not connect to database for this user: " + e.Message
    #If TargetDesktop Then
      MessageBox(msg)
    #ElseIf TargetWeb Then
      App.AppendToDBErrorLog(msg)
      App.AppendToDBErrorLog("DBFile: " + dbFile.NativePath)
    #EndIf
    Return Nil
  End Try

  Return orders
End Function</code></pre>

### UpdateInvoice

`Public`

<pre><code class="language-xojo">Function UpdateInvoice(invoiceNum As String, invoiceDate As String, invoiceTotal As String) As Boolean
  Var rs As Rowset = Self.SelectSQL("SELECT * FROM Invoices WHERE InvoiceNo=?", invoiceNum)

  If rs &lt;&gt; Nil Then
    rs.EditRow
    rs.Column("InvoiceDate").StringValue = invoiceDate
    rs.Column("InvoiceAmount").CurrencyValue = CDbl(UtilLib.Filter(invoiceTotal, "0123456789."))

    Try
      rs.SaveRow

      rs.Close
    Catch e As DatabaseException
      Return False
    End Try
  End If

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


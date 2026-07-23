---
title: App
---

# App

`Class` `Protected`

**Inherits:** [WebApplication](https://documentation.xojo.com/api/web/webapplication.html)  

## Event Handlers

### HandleURL

`Public` `Event Handler`

<pre><code>Function HandleURL(request As <a href="https://documentation.xojo.com/api/web/webrequest.html">WebRequest</a>, response As <a href="https://documentation.xojo.com/api/web/webresponse.html">WebResponse</a>) As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

https://documentation.xojo.com/topics/communication/internet/testing_a_web_service.html#testing-a-web-service

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Function HandleURL(request As WebRequest, response As WebResponse) As Boolean
      // https://documentation.xojo.com/topics/communication/internet/testing_a_web_service.html#testing-a-web-service
      Log(CurrentMethodName)
      Log(Request.Path)

      Var data As String = DefineEncoding(Request.Body, Encodings.UTF8)
      Var json As JSONItem

      Select Case Request.Path // /GetAllCustomers
      Case "GetAllCustomers"
        json = GetAllCustomers

      Case "GetCustomer"
        json = GetCustomer(data)

      Case Else
        // Do not process request
        Return False
      End Select

      // Send back data
      Response.Write(json.ToString)
      Response.Status = 200 'Let the browser know the operation was successful
      Return True
    End Function</code></pre>

</details>

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening(args() As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening(args() As String)
      #PRAGMA unused args

      // Connect to the EddiesElectronics SQLite database.
      // This can be any database or database server you want to use.

      Var dbFile As FolderItem
      dbFile = New FolderItem("EddiesElectronics.sqlite", FolderItem.PathModes.Native)

      If Not dbFile.Exists Then
        System.DebugLog("DB not found: " + dbFile.NativePath)
        Quit
      End If

      DB = New SQLiteDatabase
      DB.DatabaseFile = dbFile

      Try
        db.Connect
        System.DebugLog("Connect to database: " + dbFile.NativePath)
        Connected = True
      Catch error As DatabaseException
        System.DebugLog("Could not connect to database: " + error.Message)
        Connected = False
      End Try
    End Sub</code></pre>

</details>

## Methods

### GetAllCustomers

`Protected, Shared`

Always returns specific columns for displaying customers in a list

<pre><code class="language-xojo">Private Function GetAllCustomers() As JSONItem
  // Always returns specific columns for displaying customers in a list

  Var sql As String
  sql = "SELECT ID, FirstName, LastName, City, State, Zip FROM Customers ORDER BY FirstName"

  Var rs As RowSet
  Try
    rs = DB.SelectSQL(sql)
    Var jsonCustomers As New Dictionary
    For Each row As DatabaseRow In rs
      Var cust As New Dictionary
      cust.Value("FirstName") = row.Column("FirstName").StringValue
      cust.Value("LastName") = row.Column("LastName").StringValue
      cust.Value("City") = row.Column("City").StringValue
      cust.Value("State") = row.Column("State").StringValue
      cust.Value("Zip") = row.Column("Zip").StringValue

      jsonCustomers.Value(row.Column("ID").StringValue) = cust
    Next

    rs.Close

    Var jsonResults As New JSONItem
    jsonResults.Value("GetAllCustomers") = jsonCustomers

    Return jsonResults

  Catch error As DatabaseException
    System.DebugLog("DB Error: " + error.Message)

    Var jsonError As New JSONItem
    jsonError.Value("DBError") = error.Message
    Return jsonError
  End Try
End Function</code></pre>

### GetCustomer

`Protected, Shared`

Gets all the columns for the customer (specified by its ID)
The request supplies JSON with the ID of the customer

<pre><code class="language-xojo">Private Function GetCustomer(jsonData As String) As JSONItem
  // Gets all the columns for the customer (specified by its ID)

  // The request supplies JSON with the ID of the customer
  Var json As New JSONItem(jsonData)

  Var id As Integer = json.Value("ID")

  Var SQL As String = "SELECT * FROM Customers WHERE ID = ?"

  Var rs As RowSet

  Try
    rs = DB.SelectSQL(SQL, id)

    Var jsonCustomer As New JSONItem
    If Not rs.AfterLastRow Then
      jsonCustomer.Value("ID") = rs.Column("ID").StringValue
      jsonCustomer.Value("FirstName") = rs.Column("FirstName").StringValue
      jsonCustomer.Value("LastName") = rs.Column("LastName").StringValue
      jsonCustomer.Value("City") = rs.Column("City").StringValue
      jsonCustomer.Value("State") = rs.Column("State").StringValue
      jsonCustomer.Value("Zip") = rs.Column("Zip").StringValue
      jsonCustomer.Value("Phone") = rs.Column("Phone").StringValue
      jsonCustomer.Value("Email") = rs.Column("Email").StringValue
      jsonCustomer.Value("Photo") = EncodeBase64(rs.Column("Photo").StringValue)
      jsonCustomer.Value("Taxable") = rs.Column("Taxable").StringValue
    End If

    rs.Close

    Var jsonResults As New JSONItem
    jsonResults.Value("GetCustomer") = jsonCustomer

    Return jsonResults
  Catch error As DatabaseException
    System.DebugLog("DBError: " + error.Message)

    Var jsonError As New JSONItem
    jsonError.Value("DBError") = error.Message
    Return jsonError
  End Try
End Function</code></pre>

### GetLog

`Public`

<pre><code class="language-xojo">Function GetLog() As String
  Return mLog
End Function</code></pre>

### Log

`Public`

<pre><code class="language-xojo">Sub Log(s As String)
  mLog = s + EndOfLine + mLog
End Sub</code></pre>

## Properties

### Connected

`Public`

<pre><code>Connected As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

### mLog

`Public`

<pre><code>mLog As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### DB

`Private, Shared`

<pre><code>DB As <a href="https://documentation.xojo.com/api/databases/sqlitedatabase.html">SQLiteDatabase</a></code></pre>

</details>


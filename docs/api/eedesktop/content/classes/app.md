---
title: App
---

# App

`Class` `Protected`

**Inherits:** [DesktopApplication](https://documentation.xojo.com/api/user_interface/desktop/desktopapplication.html)  

## Event Handlers

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening()
      Orders = OrdersDatabase.SetupNewDatabase

      If Orders &lt;&gt; Nil Then
        Orders.UpdateInvoiceDates
        CustomerDetailsWindow.Show
      Else
        Quit
      End If

    End Sub</code></pre>

</details>

## Properties

### DBErrorLogFile

`Public`

<pre><code>DBErrorLogFile As <a href="https://documentation.xojo.com/api/files/folderitem.html">FolderItem</a></code></pre>

### LogDatabase

`Public`

<pre><code>LogDatabase As <a href="https://documentation.xojo.com/api/databases/sqlitedatabase.html">SQLiteDatabase</a></code></pre>

### Orders

`Public`

<pre><code>Orders As <a href="../ordersdatabase/">OrdersDatabase</a></code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### LaunchedOn

`Private, Shared`

<pre><code>LaunchedOn As <a href="https://documentation.xojo.com/api/data_types/datetime.html">DateTime</a></code></pre>

</details>

## Constants

### kEditClear

`Public`

<pre><code>Const kEditClear As String = "&amp;Delete"</code></pre>


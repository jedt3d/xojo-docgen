---
title: App
---

# App

`Class` `Protected`

**Inherits:** [WebApplication](https://documentation.xojo.com/api/web/webapplication.html)  

## Version Info

3.0 - August 2020
Converted to Web Framework 2.0 with Xojo 2020r1.
Redeveloped mobile layouts for better user experience.
Removed language localizations.

2.1 - January 2013
Added ability to add products to invoices.
Added a variety of language localizations.

2.0 - October 2012
Added OrdersDatabase subclass to shared DB code between desktop and web versions. Updated code and formattin
so that code is more similar between desktop and web versions.

1.5 - February 23rd, 2012
Added a WebButton subclass to correctly size and style buttons for Android phones and tablets
Added some more iOS styling

1.4 - November 18th, 2011
The UI now scales properly when on an iPad in portrait mode. This was done via the app.HTMLHeader property.
Added Type and Period popups to the Log page so the scope of the log view can be narrowed.

1.3 - September 2nd, 2011
Added a mobile user interface for iPhone and iPod Touch.

1.2 - August 30th, 2011
Added a toolbar and moved search, show all, revert and update buttons into the toolbar.

1.1 - August 26th, 2011
Added support for showing the customer's location in a MapViewer

1.0 August 16th, 2011
Inital release.

## Event Handlers

### HandleURL

`Public` `Event Handler`

<pre><code>Function HandleURL(request As <a href="https://documentation.xojo.com/api/web/webrequest.html">WebRequest</a>, response As <a href="https://documentation.xojo.com/api/web/webresponse.html">WebResponse</a>) As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Function HandleURL(request As WebRequest, response As WebResponse) As Boolean
      If Request.Path = "EEWeb" Or Request.Path = "EEWeb/" Then
        Response.Write "&lt;meta http-equiv=""refresh"" content=""0;URL=https://demos.xojo.com""&gt;"
        Return True
      End If
    End Function</code></pre>

</details>

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening(args() As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening(args() As String)
      #PRAGMA unused args

      #If XojoVersion &lt; 2020.01 Then
        #Pragma Error("Eddie's Electronics requires Xojo 2020r2 or later.")
      #EndIf

      OrdersDatabase.CleanDBFolder

      DBErrorLogFile = New FolderItem("DBErrorLog", FolderItem.PathModes.Native)
      LaunchedOn = DateTime.Now

      Var f As New FolderItem("Logs.sqlite", FolderItem.PathModes.Native)
      If f.Exists Then
        LogDatabase = New SQLiteDatabase
        LogDatabase.DatabaseFile = f
        If LogDatabase.Connect Then
          Var row As New DatabaseRow
          row.Column("LaunchEvent").DateTimeValue = DateTime.Now
          row.Column("Version").StringValue = App.Version

          Try
            LogDatabase.AddRow("Launch", row)
          Catch e As DatabaseException
            AppendToDBErrorLog("DBLog: " + e.Message)
          End Try
        Else
          AppendToDBErrorLog("Can't connect to database.")
        End If
      End If
    End Sub</code></pre>

</details>

### UnhandledException

`Public` `Event Handler`

<pre><code>Function UnhandledException(error As <a href="https://documentation.xojo.com/api/exceptions/runtimeexception.html">RuntimeException</a>) As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Function UnhandledException(error As RuntimeException) As Boolean
      #PRAGMA unused error

    End Function</code></pre>

</details>

## Methods

### AppendToDBErrorLog

`Public`

<pre><code class="language-xojo">Sub AppendToDBErrorLog(theError as String)
  Var tos As TextOutputStream

  If DBErrorLogFile.Exists Then
    tos = TextOutputStream.Open(DBErrorLogFile)
  Else
    tos = TextOutputStream.Create(DBErrorLogFile)
  End If

  tos.WriteLine(DateTime.Now.SQLDateTime + Chr(9) + theError)
  tos.Close
End Sub</code></pre>

### UpTime

`Public`

<pre><code class="language-xojo">Function UpTime(Optional LongFormat as Boolean = false) As String
  Var days, hours, minutes, seconds As Integer

  seconds = DateTime.Now.SecondsFrom1970 - LaunchedOn.SecondsFrom1970

  days = seconds/86400
  seconds = seconds Mod 86400

  hours = seconds / 3600
  seconds = seconds Mod 3600

  minutes = seconds/60
  seconds = seconds Mod 60

  If LongFormat Then
    Return Format(days, "000") + "days, " + Format(Hours, "00") + " hours, " + Format(Minutes, "00") + "minutes, " + Format(seconds, "00") + "seconds"
  Else
    Return Format(days, "000") + ":" + Format(Hours, "00") + ":" + Format(Minutes, "00") + ":" + Format(seconds, "00")
  End If
End Function</code></pre>

## Properties

### DBErrorLogFile

`Public`

<pre><code>DBErrorLogFile As <a href="https://documentation.xojo.com/api/files/folderitem.html">FolderItem</a></code></pre>

### LogDatabase

`Public`

<pre><code>LogDatabase As <a href="https://documentation.xojo.com/api/databases/sqlitedatabase.html">SQLiteDatabase</a></code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### LaunchedOn

`Private, Shared`

<pre><code>LaunchedOn As <a href="https://documentation.xojo.com/api/data_types/datetime.html">DateTime</a></code></pre>

### SystemUsage

`Private, Shared`

<pre><code>SystemUsage As <a href="../systemusagelogger/">SystemUsageLogger</a></code></pre>

</details>


---
title: Session
---

# Session

`Session` `Protected`

**Inherits:** [WebSession](https://documentation.xojo.com/api/web/websession.html)  

## Event Handlers

### Closing

`Public` `Event Handler`

<pre><code>Sub Closing(appQuitting As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Closing(appQuitting As Boolean)
      #PRAGMA unused appQuitting

      If Orders &lt;&gt; Nil Then
        Orders.Close
        Orders.DatabaseFile.Remove
        Orders = Nil
      End If
    End Sub</code></pre>

</details>

### HashtagChanged

`Public` `Event Handler`

<pre><code>Sub HashtagChanged(name As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>, data As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub HashtagChanged(name As String, data As String)
      Select Case Name
      Case "Log"
        LogPage.Show
      Case "customerID"
        If Data.ToDouble &gt; 0 Then
          CustomerDetailsPage.SelectCustomerByID(Data)
        End If
      Else

      End Select
    End Sub</code></pre>

</details>

### JavaScriptError

`Public` `Event Handler`

<pre><code>Sub JavaScriptError(errorName as <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>, errorMessage as <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>, errorStack as <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub JavaScriptError(errorName as String, errorMessage as String, errorStack as String)
      #PRAGMA unused ErrorName
      #PRAGMA unused ErrorMessage
      #PRAGMA unused ErrorStack

      Var row As New DatabaseRow
      row.Column("ErrorEvent") = DateTime.Now
      row.Column("UserAddress") = Session.RemoteAddress
      row.Column("UserDetails") = "User Details"
      row.Column("ErrorMessage") = ErrorMessage

      Try
        App.LogDatabase.AddRow("Errors", row)
      Catch e As DatabaseException
        App.AppendToDBErrorLog("DBLog: " + e.Message)
      End Try

    End Sub</code></pre>

</details>

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening()
      Self.ConfirmDisconnectMessage = "You are about to leave the Eddie's Electronics application."
      If Not DebugBuild Then Self.UserTimeout = 300

      // Setup a copy of the database so the user can change the data all they want
      Var db As OrdersDatabase = OrdersDatabase.SetupNewDatabase
      If db &lt;&gt; Nil Then
        Orders = db
        Connected = True
        db.UpdateInvoiceDates
      Else
        DatabaseNotAvailablePage.Show
        Return
      End If

      // Log the user accessing the app
      Var row As New DatabaseRow
      row.Column("AccessEvent").DateTimeValue = DateTime.Now
      row.Column("UserAddress").StringValue = RemoteAddress
      row.Column("SessionCount").IntegerValue = App.SessionCount + 1
      Try
        App.LogDatabase.AddRow("Access", row)
      Catch e As DatabaseException
        App.AppendToDBErrorLog("DBLog: " + e.Message)
      End Try

    End Sub</code></pre>

</details>

### UserTimedOut

`Public` `Event Handler`

<pre><code>Sub UserTimedOut()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub UserTimedOut()
      Self.ConfirmDisconnectMessage = ""
      Self.GotoURL("https://www.xojo.com/store")
    End Sub</code></pre>

</details>

## Methods

### UpdateMap

`Public`

<pre><code class="language-xojo">Sub UpdateMap(theMap As WebMapViewer, cityStateZip As String)
  #Pragma BreakOnExceptions Off
  // Update the Map
  // We are using a Try-Catch statement here because if the city, state, zip passed to WebMapLocation are not valid
  // a NilObjectException will be thrown. In that unlikely event, a MessageBox will appear as you can see below in the catch section.
  If CurrentLocation &lt;&gt; Nil Then CurrentLocation.Visible = False // If there's one already there, remove it.
  Try
    CurrentLocation = New WebMapLocation(cityStateZip + " USA")

    theMap.GoToLocation CurrentLocation // Center the map on the location
    theMap.AddLocation CurrentLocation // Add a map marker to the map at that location
    CurrentLocation.Visible = True // Show the map marker

  Catch err As NilObjectException
    // MessageBox("According to Google, this location does not exist.")
  End Try
  #Pragma BreakOnExceptions Default
End Sub</code></pre>

## Properties

### CurrentLocation

`Public`

<pre><code>CurrentLocation As <a href="https://documentation.xojo.com/api/web/webmaplocation.html">WebMapLocation</a></code></pre>

'This is stored so that we can remove the location on the map when we switch to the next customer

### MobileUser

`Public`

<pre><code>MobileUser As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

### Orders

`Public`

<pre><code>Orders As <a href="../ordersdatabase/">OrdersDatabase</a></code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### Connected

`Private, Shared`

<pre><code>Connected As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

</details>


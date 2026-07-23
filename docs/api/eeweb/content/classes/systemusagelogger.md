---
title: SystemUsageLogger
---

# SystemUsageLogger

`Class` `Protected`

**Inherits:** [Timer](https://documentation.xojo.com/api/language/timer.html)  

## Event Handlers

### Action

`Public` `Event Handler`

<pre><code>Sub Action()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Action()
      ServerStatus.Poll

      Var row As New DatabaseRow
      row.Column("ReadingEvent").DateTimeValue = DateTime.Now
      row.Column("CPU").IntegerValue = 0
      row.Column("RAM").IntegerValue = 0

      Try
        App.LogDatabase.AddRow("SystemUsage", row)
      Catch e As DatabaseException
        App.AppendToDBErrorLog("DBLog: " + e.Message)
      End Try

    End Sub</code></pre>

</details>

## Methods

### Constructor

`Public`

<pre><code class="language-xojo">Sub Constructor()
  ServerStatus = new MachineStatus
End Sub</code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### ServerStatus

`Private, Shared`

<pre><code>ServerStatus As <a href="../machinestatus/">MachineStatus</a></code></pre>

</details>


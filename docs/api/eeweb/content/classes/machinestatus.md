---
title: MachineStatus
---

# MachineStatus

`Class` `Protected`

## Methods

### Poll

`Public`

<pre><code class="language-xojo">Sub Poll()
  Var rv As String = Testdata
  #If TargetLinux Then
    If sh=Nil Then sh = New Shell

    Var cmd As String = "vmstat -a 1 1"
    sh.ExecuteMode = Shell.ExecuteModes.Synchronous
    sh.Execute cmd
    rv = sh.Result
  #EndIf

  rv = rv.ReplaceLineEndings(EndOfLine).Trim

  Var sa() As String = rv.Split(EndOfLine)

  //Remove blank lines from the end
  While sa(sa.LastIndex) = ""
    sa.RemoveAt(sa.LastIndex)
  Wend

  //Grab the last line
  Var lastline As String = sa(sa.LastIndex)
  //Remove extra spaces
  While lastline.IndexOf("  ") &gt;= 0
    lastline = lastline.ReplaceAll("  "," ").Trim
  Wend

  //Store the stats
  Var curdat() As String = lastline.Split(" ")
  //Memory
  MemoryVM = CDbl(curdat(2))
  MemoryFree = CDbl(curdat(3))
  MemoryInactive = CDbl(curdat(4))
  MemoryActive = CDbl(curdat(5))

  //System
  SystemInterrupts = CDbl(curdat(10))
  SystemContextSwitches = CDbl(curdat(11))

  //CPU
  CPUUser = CDbl(curdat(12))
  CPUSystem = CDbl(curdat(13))
  CPUIdle = CDbl(curdat(14))
  CPUWaitingForIO = CDbl(curdat(15))
End Sub</code></pre>

## Properties

### CPUIdle

`Public`

<pre><code>CPUIdle As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### CPUSystem

`Public`

<pre><code>CPUSystem As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### CPUUser

`Public`

<pre><code>CPUUser As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### CPUWaitingForIO

`Public`

<pre><code>CPUWaitingForIO As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### MemoryActive

`Public`

<pre><code>MemoryActive As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### MemoryFree

`Public`

<pre><code>MemoryFree As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### MemoryInactive

`Public`

<pre><code>MemoryInactive As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### MemoryVM

`Public`

<pre><code>MemoryVM As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### SystemContextSwitches

`Public`

<pre><code>SystemContextSwitches As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### SystemInterrupts

`Public`

<pre><code>SystemInterrupts As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### Properties — internal

<details class="internal"><summary>Private / internal members</summary>

### sh

`Private, Shared`

<pre><code>sh As <a href="https://documentation.xojo.com/api/os/shell.html">Shell</a></code></pre>

</details>

### Constants — internal

<details class="internal"><summary>Private / internal members</summary>

### Testdata

`Private`

<pre><code>Const Testdata As String = "procs -----------memory---------- ---swap-- -----io---- --system-- -----cpu------  r  b   swpd   free  inact active   si   so    bi    bo   in   cs us sy id wa st  0  0    108  91420 199936 3492860    0    0     0     1    0    0  0  0 100  0  0"</code></pre>

</details>


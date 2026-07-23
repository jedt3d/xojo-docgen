---
title: DataPoint
---

# DataPoint

`Class` `Protected`

## Methods

### Constructor

`Public`

<pre><code class="language-xojo">Sub Constructor(value as Integer, shape as Integer = 0)
  Self.Value = value
  Self.Shape = shape
End Sub</code></pre>

### PointNearby

`Public`

<pre><code class="language-xojo">Function PointNearby(x as integer, y as integer, tolerance as integer = 10) As Boolean
  If Abs(x-Self.x) &lt; tolerance And Abs(y-Self.y) &lt; tolerance Then
    Return True
  End If
End Function</code></pre>

## Properties

### Label

`Public`

<pre><code>Label As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Shape

`Public`

<pre><code>Shape As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### Value

`Public`

<pre><code>Value As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### X

`Public`

<pre><code>X As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### Y

`Public`

<pre><code>Y As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>


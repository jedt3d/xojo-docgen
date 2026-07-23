---
title: Customer
---

# Customer

`Class` `Protected`

## Methods

### Save

`Public`

<pre><code class="language-xojo">Sub Save()
  Var sql As String
  sql = "UPDATE Customers SET FirstName = ?1, LastName = ?2, " + _
  "Address = ?3, City = ?4, State = ?5, Zip = ?6, Phone = ?7, " + _
  "Email = ?8, Taxable = ?9 WHERE ID = ?10"

  // Pass in values after sql instead of doing string replacement
  App.EEDB.ExecuteSQL(sql, FirstName, LastName, Address, City, State, _
  Zip, Phone, Email, Taxable, ID)

End Sub</code></pre>

## Properties

### Address

`Public`

<pre><code>Address As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### City

`Public`

<pre><code>City As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Email

`Public`

<pre><code>Email As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### FirstName

`Public`

<pre><code>FirstName As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### ID

`Public`

<pre><code>ID As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

### LastName

`Public`

<pre><code>LastName As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Phone

`Public`

<pre><code>Phone As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Photo

`Public`

<pre><code>Photo As <a href="https://documentation.xojo.com/api/graphics/picture.html">Picture</a></code></pre>

### State

`Public`

<pre><code>State As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

### Taxable

`Public`

<pre><code>Taxable As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

### Zip

`Public`

<pre><code>Zip As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a></code></pre>

## Computed Properties

### FullAddress

`Public`

<pre><code>FullAddress As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a> (read-only)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Get
  Var addr As String = Address + EndOfLine.iOS + _
  City + ", " + State + " " + Zip

  Return addr

End Get</code></pre>

</details>

### FullName

`Public`

<pre><code>FullName As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a> (read-only)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Get
  Return FirstName + " " + LastName
End Get</code></pre>

</details>

### ShortAddress

`Public`

<pre><code>ShortAddress As <a href="https://documentation.xojo.com/api/data_types/string.html">String</a> (read-only)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Get
  Return City + ", " + State + " " + Zip
End Get</code></pre>

</details>


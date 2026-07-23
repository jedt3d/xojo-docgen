---
title: SMTP
---

# SMTP

`Class` `Protected`

**Inherits:** [SMTPSecureSocket](https://documentation.xojo.com/api/networking/smtpsecuresocket.html)  

## Event Handlers

### ConnectionEstablished

`Public` `Event Handler`

<pre><code>Sub ConnectionEstablished(greeting as string)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub ConnectionEstablished(greeting as string)
      System.DebugLog(CurrentMethodName + ": " + greeting)
    End Sub</code></pre>

</details>

### Error

`Public` `Event Handler`

<pre><code>Sub Error(err As <a href="https://documentation.xojo.com/api/exceptions/runtimeexception.html">RuntimeException</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Error(err As RuntimeException)
      #Pragma unused err

      System.DebugLog(CurrentMethodName + ": " + err.ErrorNumber.ToString)
      error = True
    End Sub</code></pre>

</details>

### MailSent

`Public` `Event Handler`

<pre><code>Sub MailSent()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub MailSent()
      Finished = True
      System.DebugLog("Mail sent")
    End Sub</code></pre>

</details>

### MessageSent

`Public` `Event Handler`

<pre><code>Sub MessageSent(Email as <a href="https://documentation.xojo.com/api/networking/emailmessage.html">EmailMessage</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub MessageSent(Email as EmailMessage)
      System.DebugLog(CurrentMethodName + ": " + email.Subject)
    End Sub</code></pre>

</details>

### SendProgress

`Public` `Event Handler`

<pre><code>Function SendProgress(BytesSent As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a>, BytesLeft As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a>) As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Function SendProgress(BytesSent As Integer, BytesLeft As Integer) As Boolean
      System.DebugLog(CurrentMethodName + ": " + Str(BytesSent) + " of " + Str(BytesSent + BytesLeft))
    End Function</code></pre>

</details>

### ServerError

`Public` `Event Handler`

<pre><code>Sub ServerError(ErrorID as integer, ErrorMessage as string, Email as <a href="https://documentation.xojo.com/api/networking/emailmessage.html">EmailMessage</a>)</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub ServerError(ErrorID as integer, ErrorMessage as string, Email as EmailMessage)
      #PRAGMA unused ErrorID
      #PRAGMA unused Email

      System.DebugLog(CurrentMethodName + ": " + ErrorMessage)
      error = True
    End Sub</code></pre>

</details>

## Properties

### Error

`Public`

<pre><code>Error As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>

### Finished

`Public`

<pre><code>Finished As <a href="https://documentation.xojo.com/api/data_types/boolean.html">Boolean</a></code></pre>


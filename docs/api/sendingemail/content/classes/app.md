---
title: App
---

# App

`Class` `Protected`

**Inherits:** [ConsoleApplication](https://documentation.xojo.com/api/console/consoleapplication.html)  

## Event Handlers

### Run

`Public` `Event Handler`

<pre><code>Function Run(args() as <a href="https://documentation.xojo.com/api/data_types/string.html">String</a>) As <a href="https://documentation.xojo.com/api/data_types/integer.html">Integer</a></code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Function Run(args() as String) As Integer
      #PRAGMA unused args

      SendMail

      // Wait for the mail to finish sending before letting the
      // Console app quit.
      While Not SendMailSocket.Finished And Not SendMailSocket.Error
        App.DoEvents
      Wend
    End Function</code></pre>

</details>

## Methods

### SendMail

`Public`

<pre><code class="language-xojo">Sub SendMail()
  SendMailSocket = New SMTP
  SendMailSocket.Address = "your.smtp.com" // your SMTP email server
  SendMailSocket.Port = 587 // Check your server for the property port # to use
  SendMailSocket.SSLConnectionType = SMTPSecureSocket.SSLConnectionTypes.TLSv1
  SendMailSocket.Username = "YourEmailUsername"
  SendMailSocket.Password = "YourEmailPassword"

  // Create the actual email message
  Var mail As New EmailMessage
  mail.FromAddress = "sender@domain.com"
  mail.Subject = "////&gt; Test Email from Xojo"
  mail.BodyPlainText = "Hello, World!"
  mail.Headers.AddHeader("X-Mailer", "Xojo SMTP Example") // Sample header
  mail.AddRecipient("recipient@domain.com")

  // Add the message to the SMTPSocket and send it
  SendMailSocket.Messages.Add(mail)
  SendMailSocket.SendMail
  System.DebugLog("SendMail: Done")
End Sub</code></pre>

## Properties

### SendMailSocket

`Public`

<pre><code>SendMailSocket As <a href="../smtp/">SMTP</a></code></pre>


---
title: CustomerDetailsToolbar
---

# CustomerDetailsToolbar

`Class` `Protected`

**Inherits:** [WebToolbar](https://documentation.xojo.com/api/user_interface/web/webtoolbar.html)  

## Event Definitions

### 

`Event Definition`

<pre><code>Event Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Event Opening()</code></pre>

</details>

## Event Handlers

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening()
      Self.Title = "Eddie՚s Electronics"
      Self.Icon = EE_icon

      Var btn As New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.Icon = WebPicture.BootstrapIcon("Search")
      btn.Caption = "Search"
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.Icon = WebPicture.BootstrapIcon("People")
      btn.Caption = "Show All"
      btn.Enabled = False
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.FlexibleSpace
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.icon = WebPicture.BootstrapIcon("Bar Chart Line")
      btn.Caption = "Sales Chart"
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.FlexibleSpace
      Self.AddItem(btn)


      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.Icon = WebPicture.BootstrapIcon("X Circle")
      btn.Caption = "Undo"
      btn.Enabled = False
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.Icon = WebPicture.BootstrapIcon("Check Circle")
      btn.Caption = "Update"
      btn.Enabled = False
      Self.AddItem(btn)

      Opening
    End Sub</code></pre>

</details>


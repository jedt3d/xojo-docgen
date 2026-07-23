---
title: MobileToolbar
---

# MobileToolbar

`Class` `Protected`

**Inherits:** [WebToolbar](https://documentation.xojo.com/api/user_interface/web/webtoolbar.html)  

## Event Handlers

### Opening

`Public` `Event Handler`

<pre><code>Sub Opening()</code></pre>

<details class="source"><summary>Source</summary>

<pre><code class="language-xojo">Sub Opening()
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
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.icon = WebPicture.BootstrapIcon("Bar Chart Line")
      btn.Caption = "Sales Chart"
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.Separator
      Self.AddItem(btn)

      btn = New WebToolbarButton
      btn.style = WebToolbarButton.ButtonStyles.PushButton
      btn.icon = EE_icon
      btn.Caption = "About..."
      Self.AddItem(btn)
    End Sub</code></pre>

</details>


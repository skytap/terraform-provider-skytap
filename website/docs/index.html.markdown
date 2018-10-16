---
layout: "skytap"
page_title: "Provider: Skytap"
sidebar_current: "docs-skytap-index"
description: |-
  The Skytap provider is used to interact with the resources supported by Skytap. 
  The provider needs to be configured with the proper credentials before it can be used.
---

# Skytap Provider

The Skytap provider is used to interact with the resources supported by Skytap. 
  The provider needs to be configured with the proper credentials before it can be used.

A typical provider configuration will look something like:

```hcl
provider "skytap" {
  username   = "xxx"
  password   = "yyy"
}
```

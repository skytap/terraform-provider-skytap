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

Use the navigation to the left to read about the available resources.

A typical provider configuration will look something like:

```hcl
provider "skytap" {
  username = "${var.skytap_username}"
  api_token = "${var.skytap_api_token}"
}

resource "skytap_environment" "env" {
	# ...
}
```

## Arguments Reference

The following arguments are supported:

* `username` - (Required) This is the Skytap username. This can also be specified
  with the `SKYTAP_USERNAME` shell environment variable.
* `api_token` - (Required) This is the Skytap API token. This can also be specified
  with the `SKYTAP_API_TOKEN` shell environment variable.




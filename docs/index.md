---
page_title: "skytap Provider"
subcategory: ""
description: |-
  The Skytap provider is used to interact with the resources supported by Skytap.
---

# skytap Provider

The Skytap provider is used to interact with the resources supported by Skytap.
The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

A typical provider configuration will look something like:

```hcl
provider "skytap" {
  username = var.skytap_username
  api_token = var.skytap_api_token
}

resource "skytap_environment" "env" {
	# ...
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **api_token** (String) The Skytap API token. May also be specified by the `SKYTAP_API_TOKEN` shell environment variable
- **username** (String) The Skytap username. May also be specified by the `SKYTAP_USERNAME` shell environment variable

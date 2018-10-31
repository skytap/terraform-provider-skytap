---
layout: "skytap"
page_title: "Skytap: skytap_network"
sidebar_current: "docs-skytap-resource-network"
description: |-
  Provides a Skytap Network resource.
---

# skytap\_network

Provides a Skytap Network resource. Networks are not top-level elements of the Skytap API. 
Rather, they are elements properly contained within an environment. 
Operations on them are implicitly on the containing environment.

## Example Usage


```hcl
# Create a new network
resource "skytap_network" "network" {
  environment_id = "123456"
  name = "my network"
  domain = "domain.com"
  subnet = "1.2.3.4/16"
  gateway = "1.2.3.254"
  tunnelable = true
}
```

## Argument Reference

The following arguments are supported:

* `environment_id` - (Required, Force New) ID of the environment you want to attach the network to. If updating with a new one then the network will be recreated.
* `name` - (Required) User-defined name of the network. Limited to 255 characters. UTF-8 character type.
* `domain` - (Optional) Domain name for the Skytap network. Limited to 64 characters.
                         <br/>Valid characters are lowercase letters, numbers, and hyphens. Cannot be blank, must not begin or end with a period, and must start and end with a letter or number.
                         <br/>This field can be changed only when all virtual machines in the environment are stopped (not suspended or running).
* `subnet` - (Optional) Defines the subnet address and subnet mask size in CIDR format (for example, 10.0.0.0/24). IP addresses for the VMs are assigned from this subnet and standard network services (DNS resolution, CIFS share, routes to Internet) are defined appropriately for it.
                        <br/>The subnet mask size must be between 16 and 29. Valid characters are lowercase letters, numbers, and hyphens. Cannot be blank, must not begin or end with a period, and must start and end with a letter or number.                
* `gateway` - (Optional, Computed) Gateway IP address.
* `tunnelable` - (Optional) If true, this network can be connected to other networks.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the network.
* `environment_id`: ID of the environment the network belongs to. 
* `name`: User-defined name of the network.
* `domain`: Domain name for the Skytap network.
* `subnet`: Defines the subnet address and subnet mask size in CIDR format.
* `gateway`: Gateway IP address.
* `tunnelable`: If true, this network can be connected to other networks.

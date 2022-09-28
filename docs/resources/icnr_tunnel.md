---
page_title: "skytap_icnr_tunnel Resource - terraform-provider-skytap"
subcategory: ""
description: |-
  Provides ICNR Tunnel resource.
---

# skytap_icnr_tunnel (Resource)

Provides ICNR Tunnel. Inter-Configuration Network Routing connects networks from different environments to a single, shared server. Without ICNR, environments are on isolated networks.

## Example Usage

```hcl
resource "skytap_environment" "env1" {
    template_id = "123"
    name = "env1"
    description = "This is an environment example"
}

resource "skytap_environment" "env2" {
    template_id = "%s"
    name = "%s-environment-%d"
    description = "This is an environment example"
}

resource "skytap_network" "net1" {
    environment_id = skytap_environment.env1.id
    name = "net1"
    domain = "domain.com"
    subnet = "10.0.100.0/24"
    gateway = "10.0.100.254"
    tunnelable = true
}

resource "skytap_network" "net2" {
    environment_id = skytap_environment.env2.id
    name = "net2"
    domain = "domain.com"
    subnet = "10.0.200.0/24"
    gateway = "10.0.200.254"
    tunnelable = true
}

# Create an ICNR Tunnel between both networks.
resource "skytap_icnr_tunnel" "tunnel" {
    source = skytap_network.net1.id
    target = skytap_network.net2.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **source** (Number) Source network from where the connection is initiated. This network does not need to be 'tunnelable' (visible to other networks)
- **target** (Number) Target network to which the connection is made. The network does need to be 'tunnelable' (visible to other networks)

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)
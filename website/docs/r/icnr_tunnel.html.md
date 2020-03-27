---
layout: "skytap"
page_title: "Skytap: skytap_label_category"
sidebar_current: "docs-skytap-resource-icnr_tunnel"
description: |-
  Provides ICNR Tunnel resource.
---

# skytap\_icnr\_tunnel

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
    environment_id = "${skytap_environment.env1.id}"
    name = "net1"
    domain = "domain.com"
    subnet = "10.0.100.0/24"
    gateway = "10.0.100.254"
    tunnelable = true
}

resource "skytap_network" "net2" {
    environment_id = "${skytap_environment.env2.id}"
    name = "net2"
    domain = "domain.com"
    subnet = "10.0.200.0/24"
    gateway = "10.0.200.254"
    tunnelable = true
}

# Create an ICNR Tunnel between both networks.
resource "skytap_icnr_tunnel" "tunnel" {
    source = "${skytap_network.net1.id}"
    target = "${skytap_network.net2.id}"
}
```


## Argument Reference

The following arguments are supported:

* `source` - (Required, Force New) Source network from where the connection was initiated. This network does not need to be “tunnelable” (visible to other networks).
* `target` - (Required, Force New) Target network to which the connection was made. The network needs to be “tunnelable” (visible to other networks)

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when creating the tunnel
* `update` - (Defaults to 10 mins) Used when updating the tunnel
* `delete` - (Defaults to 10 mins) Used when destroying the tunnel

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the tunnel.

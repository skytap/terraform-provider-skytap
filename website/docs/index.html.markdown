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

Use the navigation to the left to read about the available data sources.

## Example Usage

```hcl
# Template for initial configuration bash script
data "template_file" "init" {
  template = "${file("init.tpl")}"

  vars {
    consul_address = "${aws_instance.consul.private_ip}"
  }
}

# Create a web server
resource "aws_instance" "web" {
  # ...

  user_data = "${data.template_file.init.rendered}"
}
```

Or using an inline template:

```hcl
# Template for initial configuration bash script
data "template_file" "init" {
  template = "$${consul_address}:1234"

  vars {
    consul_address = "${aws_instance.consul.private_ip}"
  }
}

# Create a web server
resource "aws_instance" "web" {
  # ...

  user_data = "${data.template_file.init.rendered}"
}
```

-> **Note:** Inline templates must escape their interpolations (as seen
by the double `$` above). Unescaped interpolations will be processed
_before_ the template.

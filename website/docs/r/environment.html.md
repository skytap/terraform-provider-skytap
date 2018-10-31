---
layout: "skytap"
page_title: "Skytap: skytap_environment"
sidebar_current: "docs-skytap-resource-environment"
description: |-
  Provides a Skytap Environment resource.
---

# skytap\_environment

Provides a Skytap Environment resource. An environment consists of one or more virtual machines, networks, 
and associated settings and metadata. Unlike a template, an environment can be run and have most of its settings 
modified. When an environment is created all of its VMs will be run.

## Example Usage


```hcl
# Create a new environment
resource "skytap_environment" "environment" {
  template_id = "123456"
  name = "Terraform Example"
  description = "Skytap terraform provider example environment."
}
```

## Argument Reference

The following arguments are supported:

* `template_id` - (Required, Force New) ID of the template you want to create an environment from. If updating with a new one then the environment will be recreated.
* `project_id` - (Optional) ID of the project you want to add the environment to. Will have no affect when updating.
* `name` - (Optional) User-defined name of the environment. Limited to 255 characters. UTF-8 character type. Will default to source template’s name if null is provided.
* `description` - (Optional) User-defined description of the environment. Limited to 1000 characters. Null allowed. UTF-8 character type.
* `outbound_traffic` - (Optional) Indicates whether networks in the environment can send outbound traffic.
* `routable` - (Optional) Indicates whether networks within the environment can route traffic to one another.
* `suspend_on_idle` - (Optional) The number of seconds an environment can be idle before it is automatically suspended.
                                 <br/>If suspend_on_idle and suspend_at_time are both null, automatic suspend is disabled.
* `suspend_at_time` - (Optional) The date and time that the environment will be automatically suspended.
                                 <br/>If suspend_on_idle and suspend_at_time are both null, automatic suspend is disabled.
* `shutdown_on_idle` - (Optional) The number of seconds an environment can be idle before it is automatically shut down.
                                  <br/>If shutdown_on_idle and shutdown_at_time are both null, automatic shut down is disabled.
* `shutdown_at_time` - (Optional) The date and time that the environment will be automatically shut down.
                                  <br/>If shutdown_on_idle and shutdown_at_time are both null, automatic shut down is disabled.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the environment.
* `name`: User-defined name of the environment.
* `description`: Summary description of the environment.
* `outbound_traffic`: Indicates whether networks in the environment can send outbound traffic.
* `routable`: Indicates whether networks within the environment can route traffic to one another.
* `suspend_on_idle`: The number of seconds an environment can be idle before it is automatically suspended.
* `suspend_at_time`: The date and time that the environment will be automatically suspended.
* `shutdown_on_idle`: The number of seconds an environment can be idle before it is automatically shut down.
* `shutdown_at_time`: The date and time that the environment will be automatically shut down.

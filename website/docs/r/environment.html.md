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
* `name` - (Required) User-defined name of the environment. Limited to 255 characters. UTF-8 character type. Will default to source template’s name if null is provided.
* `description` - (Required) User-defined description of the environment. Limited to 1000 characters. Null allowed. UTF-8 character type.
* `outbound_traffic` - (Optional) Indicates whether networks in the environment can send outbound traffic.
* `routable` - (Optional) Indicates whether networks within the environment can route traffic to one another.
* `suspend_on_idle` - (Optional) The number of seconds an environment can be idle before it is automatically suspended. Valid range: 300 to 86400 seconds (5 minutes to 1 day).
* `suspend_at_time` - (Optional) The date and time that the environment will be automatically suspended. Format: yyyy/mm/dd hh:mm:ss. By default, the suspend time uses the UTC offset for the time zone defined in your user account settings. Optionally, a different UTC offset can be supplied (for example: 2018/07/20 14:20:00 -0000). The value in the API response is converted to your time zone.
* `shutdown_on_idle` - (Optional) The number of seconds an environment can be idle before it is automatically shut down. Valid range: 300 to 86400 seconds (5 minutes to 1 day).
* `shutdown_at_time` - (Optional) The date and time that the environment will be automatically shut down. Format: yyyy/mm/dd hh:mm:ss. By default, the suspend time uses the UTC offset for the time zone defined in your user account settings. Optionally, a different UTC offset can be supplied (for example: 2018/07/20 14:20:00 -0000). The value in the API response is converted to your time zone.

~> **NOTE:**
* If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled.
* If `shutdown_on_idle` and `shutdown_at_time` are both null, automatic shut down is disabled.
* An environment cannot be set to automatically suspend and shut down. Only one of the following settings can take effect: `suspend_on_idle`, `suspend_at_time`, `shutdown_on_idle`, or `shutdown_at_time`.
* When you send a request that updates one of the four suspend or shutdown options, the other three options are automatically set to null by the REST API.
* If multiple suspend or shut down options are sent in the same request, the `suspend_type` field determines which setting Skytap Cloud will honor.
## Attributes Reference

The following attributes are exported:

* `id`: The ID of the environment.

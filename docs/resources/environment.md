---
page_title: "skytap_environment Resource - terraform-provider-skytap"
subcategory: ""
description: |-
  Provides a Skytap Environment resource.
---

# skytap_environment (Resource)

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
  tags = ["example"]
}
```

~> **NOTE:** If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled.

~> **NOTE:** If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled. If `shutdown_on_idle` and `shutdown_at_time` are both null, automatic shut down is disabled.

~> **NOTE:** If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled.* An environment cannot be set to automatically suspend and shut down. Only one of the following settings can take effect: `suspend_on_idle`, `suspend_at_time`, `shutdown_on_idle`, or `shutdown_at_time`.

~> **NOTE:** If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled. When you send a request that updates one of the four suspend or shutdown options, the other three options are automatically set to null by the REST API.

~> **NOTE:** If `suspend_on_idle` and `suspend_at_time` are both null, automatic suspend is disabled. If multiple suspend or shut down options are sent in the same request, the `suspend_type` field determines which setting Skytap Cloud will honor.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **description** (String) User-defined description of the environment. Limited to 1000 characters. UTF-8 character type
- **name** (String) User-defined name of the environment. Limited to 255 characters. UTF-8 character type
- **template_id** (String) ID of the template you want to create the environment from. If updated with a new ID, the environment will be recreated

### Optional

- **id** (String) The ID of this resource.
- **label** (Block Set) Set of labels for the instance (see [below for nested schema](#nestedblock--label))
- **disable_internet** (Boolean) Indicates whether networks in the environment allow outbound internet traffic
- **outbound_traffic** (Boolean) **DEPRECATED** Indicates whether networks in the environment can send outbound traffic. Use `disable_internet` instead
- **routable** (Boolean) Indicates whether networks within the environment can route traffic to one another
- **shutdown_at_time** (String) The date and time that the environment will be automatically shut down. Format: yyyy/mm/dd hh:mm:ss. By default, the suspend time uses the UTC offset for the time zone defined in your user account settings. Optionally, a different UTC offset can be supplied (for example: 2018/07/20 14:20:00 -0000). The value in the API response is converted to your time zone
- **shutdown_on_idle** (Number) The number of seconds an environment can be idle before it is automatically shut down. Valid range: 300 to 86400 seconds (5 minutes to 1 day)
- **suspend_at_time** (String) The date and time that the environment will be automatically suspended. Format: yyyy/mm/dd hh:mm:ss. By default, the suspend time uses the UTC offset for the time zone defined in your user account settings. Optionally, a different UTC offset can be supplied (for example: 2018/07/20 14:20:00 -0000). The value in the API response is converted to your time zone
- **suspend_on_idle** (Number) The number of seconds an environment can be idle before it is automatically suspended. Valid range: 300 to 86400 seconds (5 minutes to 1 day)
- **tags** (Set of String) Set of environment tags
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **user_data** (String) Environment user data, available from the metadata server and the Skytap API

<a id="nestedblock--label"></a>
### Nested Schema for `label`

Required:

- **category** (String) Label category that provides contextual meaning
- **value** (String) Label valueto be used for reporting

Read-Only:

- **id** (String) The ID of this resource.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String)
- **delete** (String)
- **update** (String)

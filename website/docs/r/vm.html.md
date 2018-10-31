---
layout: "skytap"
page_title: "Skytap: skytap_vm"
sidebar_current: "docs-skytap-resource-vm"
description: |-
  Provides a Skytap Virtual Machine (VM) resource.
---

# skytap\_vm

Provides a Skytap Virtual Machine (VM) resource. The environment VM resource represents an image of a single virtual machine.
<br/>Note that:
* VMs do not exist outside of environments or templates.
* An environment or template can have multiple VMs.
* Each VM is a unique resource. Therefore, a VM in a template will have a different ID than a VM in an environment created from that template.
* The VM will be run immediately after creation.

## Example Usage


```hcl
# Create a new vm
resource "skytap_vm" "vm" {
  template_id = 123
  vm_id = 456
  environment_id = 789
  name = "my vm"
}
```

## Argument Reference

The following arguments are supported:

* `template_id` - (Required, Force New) ID of the template you want to create the vm from. If updating with a new one then the VM will be recreated.
* `vm_id` - (Required, Force New) ID of the VM you want to create the VM from. If updating with a new one then the VM will be recreated.
* `environment_id` - (Required, Force New) ID of the environment you want to add the VM to. If updating with a new one then the VM will be recreated.
* `name` - (Optional, Computed) User-defined name. Limited to 100 characters. 
<br/>The name will be truncated to 33 UTF-8 characters after saving. 
<br/>If a name is not provided then the source VM's name will be used.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the VM.
* `environment_id`: ID of the environment the VM belongs to. 
* `name`: User-defined name of the VM.

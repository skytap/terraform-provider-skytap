---
layout: "skytap"
page_title: "Skytap: skytap_vm"
sidebar_current: "docs-skytap-resource-vm"
description: |-
  Provides a Skytap Virtual Machine (VM) resource.
---

# skytap\_vm

Provides a Skytap Virtual Machine (VM) resource. The environment VM resource represents an image of a single virtual machine.

~> **NOTE:**
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
  cpu = 4
  ram = 4096
}
```

## Argument Reference

The following arguments are supported:

* `environment_id` - (Required, Force New) ID of the environment you want to add the VM to. If updating with a new one then the VM will be recreated.
* `template_id` - (Required, Force New) ID of the template you want to create the vm from. If updating with a new one then the VM will be recreated.
* `vm_id` - (Required, Force New) ID of the VM you want to create the VM from. If updating with a new one then the VM will be recreated.
* `name` - (Optional, Computed) User-defined name. Limited to 100 characters. 
* `cpus` - (Optional, Computed) Number of CPUs allocated to this virtual machine. Valid range is 1 to 12. Maximum limit depends on the `max_cpus` setting.
* `ram` - (Optional, Computed) Amount of RAM allocated to this VM. Valid range is 256 and 131,072 (MB). Maximum limit depends on `max_ram` setting.
* `disk` - (Optional) Array of virtual disks within the VM. The disk size is in MiB; it will be converted to GiB in the Skytap UI. The maximum disk size is 2,096,128 MiB (1.999 TiB).

  * `name` - (Required) A unique name for the disk.
  * `id` - (Computed) The ID for the disk.
  * `size` - (Required) Specify the size of the disk. The new disk’s size should be provided in MiB. The minimum disk size is 2048 MiB; the maximum is 2096128 MiB (1.999 TiB).
  * `type` - (Computed) The type of disk.
  * `controller` - (Computed) The disk controller.
  * `lun` - (Computed) The logical unit number of the disk (LUN).

  ~> **NOTE:** The name will be truncated to 33 UTF-8 characters after saving. If a name is not provided then the source VM's name will be used.
* `network_interface` - (Optional, Computed, ForceNew) A Skytap network adapter is a virtualized network interface card (also known as a network adapter). It is logically contained in a VM and attached to a network.

  * `interface_type` - (Required, Force New) Type of network that this network adapter is attached to.
  * `network_id` - (Required, Force New) ID of the network that this network adapter is attached to.
  *	`ip` - (Required, Force New) The IP address (for example, 10.1.0.37). Skytap will not assign the same IP address to multiple interfaces on the same network.
  * `hostname` - (Required, Force New) Limited to 32 characters. Valid characters are lowercase letters, numbers, and hyphens. Cannot begin or end with hyphens. Cannot be `gw`.
* `published_service` - (Optional, Force New) Generally, a published service represents a binding of a port on a network interface to an IP and port that is routable and accessible from the public Internet. This mechanism is used to selectively expose ports on the guest to the public Internet.

  ~> **NOTE:** Published services exist and are managed as aspects of network interfaces—that is, as part of the overall environment element.
  * `internal_port` - (Required, Force New) The port that is exposed on the interface. Typically this will be dictated by standard usage (e.g., port 80 for http traffic, port 22 for SSH).

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the VM.
* `published_service`: The published services.
  * `id`: The published service's ID.
  * `external_id`: The published service's external ID.
  * `external_port`: Each published service's external port.

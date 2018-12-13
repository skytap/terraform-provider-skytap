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
resource "skytap_vm" "my_vm" {
  template_id = 123
  vm_id = 456
  environment_id = 789
  name = "my vm"
  
  network_interface = {
       interface_type = "vmxnet3"
       network_id = "${skytap_network.my_network.id}"
       ip = "10.0.0.1"
       hostname = "myhost"
    
       published_service = {
          internal_port = 80
       }
   }
}
output "external_svc_ip" {
    value = "${skytap_vm.my_vm.external_ips.10-0-0-1_80}"
}
output "external_svc_port" {
    value = "${skytap_vm.my_vm.external_ports.10-0-0-1_80}"
}

```

## Argument Reference

The following arguments are supported:

* `environment_id` - (Required, Force New) ID of the environment you want to add the VM to. If updating with a new one then the VM will be recreated.
* `template_id` - (Required, Force New) ID of the template you want to create the vm from. If updating with a new one then the VM will be recreated.
* `vm_id` - (Required, Force New) ID of the VM you want to create the VM from. If updating with a new one then the VM will be recreated.
* `name` - (Optional, Computed) User-defined name. Limited to 100 characters. 

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
* `external_ips`: A map of external IP addresses. The key is the internal IP address together with the internal port number - as defined in the `network_interface` block.
* `external_ports`: A map of external ports. The key is the internal IP address together with the internal port number - as defined in the `network_interface` block.

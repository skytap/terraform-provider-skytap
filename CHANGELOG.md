## 0.9.1 (Unreleased)

FEATURES:

* `resource/skytap_vm` : Add support to change CPU and RAM settings. ([#3](https://github.com/terraform-providers/terraform-provider-skytap/issues/3), [#9](https://github.com/terraform-providers/terraform-provider-skytap/issues/9))
* `resource/skytap_vm` : Add support to add, remove and change disks. ([#2](https://github.com/terraform-providers/terraform-provider-skytap/issues/2), [#11](https://github.com/terraform-providers/terraform-provider-skytap/issues/11))
* `resource/skytap_vm` : Allow the external ip and port values of a published service to be accessed via the published service name. ([#7](https://github.com/terraform-providers/terraform-provider-skytap/issues/7))

IMPROVEMENTS:

* `resource/skytap_vm` : Retry on 422 error when creating ([#18](https://github.com/terraform-providers/terraform-provider-skytap/issues/18))
* Improve logging ([#18](https://github.com/terraform-providers/terraform-provider-skytap/issues/18))

BIG FIXES:

* `resource/skytap_vm` : Fix documentation on published services. ([#4](https://github.com/terraform-providers/terraform-provider-skytap/issues/4))
* `resource/skytap_project` : Fix documentation on auto add roles. ([#12](https://github.com/terraform-providers/terraform-provider-skytap/issues/12))
* Documentation note blocks not paginated correctly. ([#1](https://github.com/terraform-providers/terraform-provider-skytap/issues/1))

## 0.9.0 (December 06, 2018)

FEATURES:

* **New Resource:** `skytap_environment` : Allow the provisioning of a new environment from a template. ([environment resource](https://github.com/terraform-providers/terraform-provider-skytap/commit/b8659204298067bbdbc5def7a408328f6ed324b4))
* **New Resource:** `skytap_network` : Allow the provisioning of networks under an environment resource. ([network resource](https://github.com/terraform-providers/terraform-provider-skytap/commit/f89b1aa1a04d7fa09c640ab973403870cab8574d))
* **New Resource:** `skytap_vm` : Allow the creation of a virtual machine (VM) and the configuration of the network interface. ([vm resource](https://github.com/terraform-providers/terraform-provider-skytap/commit/19b03ef4c7c55cfb7765fd357668f266e6714ebc))
* **New Resource:** `skytap_project` : Allow the creation of a new project. ([project resource](https://github.com/terraform-providers/terraform-provider-skytap/commit/8b22ac59a4cf619a7b692d7b10d5886cd9cbf3e8))
* **New Datasource:** `skytap_template` : Query a template by name for use when creating an environment. ([template datasource](https://github.com/terraform-providers/terraform-provider-skytap/commit/ec560944d0765daf8399f65949fd0b1879a11275))
* **New Datasource:** `skytap_project` : Query a project by name. ([project datasource](https://github.com/terraform-providers/terraform-provider-skytap/commit/8b22ac59a4cf619a7b692d7b10d5886cd9cbf3e8))

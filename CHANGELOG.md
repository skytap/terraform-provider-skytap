## 0.9.0 (December 06, 2018)

FEATURES:

* **New Resource:** `skytap_environment` : Allow the provisioning of a new environment from a template. ([#1](https://github.com/terraform-providers/terraform-provider-skytap/issues/1))
* **New Resource:** `skytap_network` : Allow the provisioning of networks under an environment resource. ([#5](https://github.com/terraform-providers/terraform-provider-skytap/issues/5))
* **New Resource:** `skytap_vm` : Allow the creation of a virtual machine (VM) and the configuration of the network interface. ([#4](https://github.com/terraform-providers/terraform-provider-skytap/issues/4))
* **New Resource:** `skytap_project` : Allow the creation of a new project. ([#6](https://github.com/terraform-providers/terraform-provider-skytap/issues/6))
* **New Datasource:** `skytap_template` : Query a template by name for use when creating an environment. ([#2](https://github.com/terraform-providers/terraform-provider-skytap/issues/2))
* **New Datasource:** `skytap_project` : Query a project by name. ([Refactor of the base project structure](https://github.com/terraform-providers/terraform-provider-skytap/commit/8b22ac59a4cf619a7b692d7b10d5886cd9cbf3e8))

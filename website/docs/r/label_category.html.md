---
layout: "skytap"
page_title: "Skytap: skytap_label_category"
sidebar_current: "docs-skytap-resource-label_category"
description: |-
  Provides a Skytap Label Category resource.
---

# skytap\_label\_category

Provides a Skytap label category resource. Label categories provide a taxonomy of the usa reporting, once the label category is define it can be use to label resources. There are [restrictions for the use of labels](https://help.skytap.com/using-labels-for-in-depth-reporting.html#Restrictions) for example:

* An account can have a maximum of 100 active categories.
* An account can have a maximum combined total active and inactive (deleted) 200 categories.
 

~> **NOTE:** Creating a label category fail when we reuse the name and change for label single value. 

## Example Usage


```hcl
# Create a new label_category
resource "label_category" "env" {
  name = "environment"
  single_value = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) User-defined label category name.
* `single_value` - (Required) With single value labels can have only one value for a category, with false labels can have multiple values.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the label category.

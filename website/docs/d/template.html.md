---
layout: "skytap"
page_title: "Skytap: skytap_template"
sidebar_current: "docs-skytap-datasource-template"
description: |-
  Get information on a template.
---

# skytap_template

Get information on a template. This data source provides the id and name of a template as configured on your Skytap account.
This is useful in order to retrieve a template's id via its name. The name field takes a regular expression to facilitate
the matching process.

An error is triggered if:
 1. No templates can be retrieved.
 2. The template does not exist.
 3. More than one template matches the name and the `most_recent` flag is not set.
 
If more than one templates are retrieved the `most_recent` can be set. 
This will sort the results in descending order according to the creation date. The newest template will be used.

## Example Usage

Get the template:

```hcl
data "skytap_template" "example" {
 	name = "18.04"
    most_recent = true
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A regular expression on the name of a template.
* `most_recent` - (Optional) Use the most recently created template from the returned list.

## Attributes Reference

The following attributes are exported:

* `id`: The ID of the template.
* `name`: The name of the template.

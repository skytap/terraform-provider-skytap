provider "skytap" {}

resource "skytap_project" "terraform-test-project" {
  name = "terrform-test"
  summary = "This is a project created by the skytab terraform provider"
}


module github.com/terraform-providers/terraform-provider-skytap

replace github.com/skytap/skytap-sdk-go => /Users/pegerto/work/src/github.com/skytap/skytap-sdk-go

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/hashicorp/terraform v0.12.18 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.1.0
	github.com/skytap/skytap-sdk-go v0.0.0-20200116115251-aabcf4287354
	github.com/stretchr/testify v1.3.0
)

go 1.13

package skytap

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func validateNICType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(skytap.NICTypeDefault),
		string(skytap.NICTypeE1000),
		string(skytap.NICTypeE1000E),
		string(skytap.NICTypePCNet32),
		string(skytap.NICTypeVMXNet),
		string(skytap.NICTypeVMXNet3),
	}, false)
}

func validateRoleType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		string(skytap.ProjectRoleViewer),
		string(skytap.ProjectRoleParticipant),
		string(skytap.ProjectRoleEditor),
		string(skytap.ProjectRoleManager),
	}, false)
}

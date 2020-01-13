package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/skytap/skytap-sdk-go/skytap"
	"strings"
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

func validateNoSubString(subString string) schema.SchemaValidateFunc{
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		if strings.Contains(v, subString) {
			es = append(es, fmt.Errorf("property value %s contains invalid substring '%s'",v, subString))
		}
		return
	}
}

func validateNoStartWith(subString string) schema.SchemaValidateFunc{
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		if strings.HasPrefix(v, subString) {
			es = append(es, fmt.Errorf("property value %s can start with '%s'",v, subString))
		}
		return
	}
}
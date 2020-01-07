package skytap

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func TestValidateNICType(t *testing.T) {
	x := []StringValidationTestCase{
		// No errors
		{TestName: "default", Value: string(skytap.NICTypeDefault)},
		{TestName: "pcnet32", Value: string(skytap.NICTypePCNet32)},
		{TestName: "e1000", Value: string(skytap.NICTypeE1000)},
		{TestName: "e1000e", Value: string(skytap.NICTypeE1000E)},
		{TestName: "vmxnet", Value: string(skytap.NICTypeVMXNet)},
		{TestName: "vmxnet3", Value: string(skytap.NICTypeVMXNet3)},

		// With errors
		{TestName: "empty", Value: "", ExpectError: true},
		{TestName: "unexpected", Value: "Foobar", ExpectError: true},
	}

	es := testStringValidationCases(x, validateNICType())
	if len(es) > 0 {
		t.Errorf("Failed to validate NIC types: %v", es)
	}
}

func TestValidateRoleType(t *testing.T) {
	x := []StringValidationTestCase{
		// No errors
		{TestName: "editor", Value: string(skytap.ProjectRoleEditor)},
		{TestName: "manager", Value: string(skytap.ProjectRoleManager)},
		{TestName: "participant", Value: string(skytap.ProjectRoleParticipant)},
		{TestName: "viewer", Value: string(skytap.ProjectRoleViewer)},

		// With errors
		{TestName: "empty", Value: "", ExpectError: true},
		{TestName: "unexpected", Value: "Foobar", ExpectError: true},
	}

	es := testStringValidationCases(x, validateRoleType())
	if len(es) > 0 {
		t.Errorf("Failed to validate project role types: %v", es)
	}
}

type StringValidationTestCase struct {
	TestName    string
	Value       string
	ExpectError bool
}

func testStringValidationCases(cases []StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	es := make([]error, 0)
	for _, c := range cases {
		es = append(es, testStringValidation(c, validationFunc)...)
	}

	return es
}

func testStringValidation(testCase StringValidationTestCase, validationFunc schema.SchemaValidateFunc) []error {
	_, es := validationFunc(testCase.Value, testCase.TestName)
	if testCase.ExpectError {
		if len(es) > 0 {
			return nil
		}
		return []error{fmt.Errorf("didn't see expected error in case \"%s\" with string \"%s\"", testCase.TestName, testCase.Value)}
	}

	return es
}

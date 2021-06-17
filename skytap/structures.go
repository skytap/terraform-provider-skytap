package skytap

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/skytap/skytap-sdk-go/skytap"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/hashcode"
)

func flattenNetworkInterfaces(interfaces []skytap.Interface) []interface{} {
	results := make([]interface{}, 0)

	for _, v := range interfaces {
		results = append(results, flattenNetworkInterface(v))
	}

	return results
}

func flattenNetworkInterface(v skytap.Interface) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = *v.ID
	result["interface_type"] = string(*v.NICType)
	if v.NetworkID != nil {
		result["network_id"] = *v.NetworkID
	}
	if v.IP != nil {
		result["ip"] = *v.IP
	}
	if v.Hostname != nil {
		result["hostname"] = *v.Hostname
	}
	if len(v.Services) > 0 {
		result["published_service"] = flattenPublishedServices(v.Services)
	}
	return result
}

func flattenPublishedServices(services []skytap.PublishedService) []interface{} {
	results := make([]interface{}, 0)

	for _, v := range services {
		results = append(results, flattenPublishedService(v))
	}

	return results
}

func flattenPublishedService(v skytap.PublishedService) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = *v.ID
	result["internal_port"] = *v.InternalPort
	result["external_ip"] = *v.ExternalIP
	result["external_port"] = *v.ExternalPort
	if v.Name != nil {
		result["name"] = *v.Name
	}
	return result
}

func flattenDisks(disks []skytap.Disk) []interface{} {
	results := make([]interface{}, 0)

	for _, v := range disks {
		// ignore os disk for now
		if "0" != *v.LUN {
			results = append(results, flattenDisk(v))
		}
	}

	return results
}

func flattenDisk(v skytap.Disk) map[string]interface{} {
	result := make(map[string]interface{})
	size := *v.Size
	result["id"] = *v.ID
	result["size"] = size
	result["type"] = *v.Type
	result["controller"] = *v.Controller
	result["lun"] = *v.LUN
	if v.Name != nil {
		result["name"] = *v.Name
	}
	return result
}

func flattenTags(tags []skytap.Tag) []interface{} {
	flatted := make([]interface{}, len(tags))
	for i, v := range tags {
		flatted[i] = v.Value
	}
	return flatted
}

func flattenLabels(labels []*skytap.Label) []interface{} {
	flatted := make([]interface{}, len(labels))
	for i, v := range labels {
		var label = map[string]interface{}{
			"id":       v.ID,
			"category": v.LabelCategory,
			"value":    v.Value,
		}
		flatted[i] = label
	}
	return flatted
}

func flattenProjectIDs(projects []skytap.Project) []interface{} {
	flattened := make([]interface{}, len(projects))
	for i, v := range projects {
		flattened[i] = v.ID
	}
	return flattened
}

func flattenProjectEnvironments(environments []skytap.ProjectEnvironment) []interface{} {
	flattened := make([]interface{}, len(environments))
	for i, v := range environments {
		flattened[i] = v.ID
	}
	return flattened
}

func getVMNetworkInterface(id string, vm *skytap.VM) (*skytap.Interface, error) {
	for _, networkInterface := range vm.Interfaces {
		if *networkInterface.ID == id {
			return &networkInterface, nil
		}
	}
	return nil, fmt.Errorf("could not find network interface (%s) in the VM", id)
}

func buildServices(interfaces []interface{}) (map[string]int, map[string]string) {
	ports := make(map[string]int)
	ips := make(map[string]string)

	for _, v := range interfaces {
		networkInterface := v.(map[string]interface{})
		if _, ok := networkInterface["published_service"]; ok {
			publishedServices := networkInterface["published_service"].([]interface{})
			for _, v := range publishedServices {
				publishedService := v.(map[string]interface{})
				// check if terraform is managing the published services
				if publishedService["name"] != nil {
					ips[publishedService["name"].(string)] = publishedService["external_ip"].(string)
					ports[publishedService["name"].(string)] = publishedService["external_port"].(int)
				}
			}
		}
	}
	return ports, ips
}

// caseInsensitiveSuppress is a helper function to suppress property changes
func caseInsensitiveSuppress(_, old, new string, _ *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// stringCaseSensitiveHash
func stringCaseSensitiveHash(v interface{}) int {
	return hashcode.String(strings.ToLower(v.(string)))
}

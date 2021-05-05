package skytap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/skytap/skytap-sdk-go/skytap"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/hashcode"
)

func flattenNetworkInterfaces(interfaces []skytap.Interface) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range interfaces {
		results = append(results, flattenNetworkInterface(v))
	}

	return schema.NewSet(networkInterfaceHash, results)
}

func flattenNetworkInterface(v skytap.Interface) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = *v.ID
	result["interface_type"] = string(*v.NICType)
	result["network_id"] = *v.NetworkID
	result["ip"] = *v.IP
	result["hostname"] = *v.Hostname
	if len(v.Services) > 0 {
		result["published_service"] = flattenPublishedServices(v.Services)
	}
	return result
}

func flattenPublishedServices(services []skytap.PublishedService) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range services {
		results = append(results, flattenPublishedService(v))
	}

	return schema.NewSet(publishedServiceHash, results)
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

func flattenDisks(disks []skytap.Disk) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range disks {
		// ignore os disk for now
		if "0" != *v.LUN {
			results = append(results, flattenDisk(v))
		}
	}

	return schema.NewSet(diskHash, results)
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

func getVMNetworkInterface(id string, vm *skytap.VM) (*skytap.Interface, error) {
	for _, networkInterface := range vm.Interfaces {
		if *networkInterface.ID == id {
			return &networkInterface, nil
		}
	}
	return nil, fmt.Errorf("could not find network interface (%s) in the VM", id)
}

func buildServices(interfaces *schema.Set) (map[string]int, map[string]string) {
	ports := make(map[string]int)
	ips := make(map[string]string)

	for _, v := range interfaces.List() {
		networkInterface := v.(map[string]interface{})
		if _, ok := networkInterface["published_service"]; ok {
			publishedServices := networkInterface["published_service"].(*schema.Set)
			for _, v := range publishedServices.List() {
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

// Assemble the hash for the network TypeSet attribute.
func networkInterfaceHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["interface_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["network_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["ip"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["hostname"].(string)))
	if d, ok := m["published_service"]; ok {
		publishedServices := d.(*schema.Set).List()
		for _, e := range publishedServices {
			buf.WriteString(fmt.Sprintf("%d-", publishedServiceHash(e)))
		}
	}

	return hashcode.String(buf.String())
}

// Assemble the hash for the published services TypeSet attribute.
func publishedServiceHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["internal_port"].(int)))
	if d, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", d.(string)))
	}
	return hashcode.String(buf.String())
}

// Assemble the hash for the disk TypeSet attribute.
func diskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["size"].(int)))
	if d, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", d.(string)))
	}
	return hashcode.String(buf.String())
}

// caseInsensitiveSuppress is a helper function to suppress property changes
func caseInsensitiveSuppress(k, old, new string, d *schema.ResourceData) bool {
	return strings.ToLower(old) == strings.ToLower(new)
}

// stringCaseSensitiveHash
func stringCaseSensitiveHash(v interface{}) int {
	return hashcode.String(strings.ToLower(v.(string)))
}

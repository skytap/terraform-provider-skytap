package skytap

import "github.com/skytap/skytap-sdk-go/skytap"

func flattenInterfaces(interfaces []skytap.Interface) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	for _, v := range interfaces {
		result := make(map[string]interface{})
		result["interface_type"] = *v.NICType
		if v.NetworkID != nil {
			result["network_id"] = *v.NetworkID
		}
		if v.IP != nil {
			result["ip"] = *v.IP
		}
		if v.Hostname != nil {
			result["hostname"] = *v.Hostname
		}
		result["published_service"] = flattenPublishedServices(v.Services)

		results = append(results, result)
	}

	return results
}

func flattenPublishedServices(publishedServices []skytap.PublishedService) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	for _, v := range publishedServices {
		result := make(map[string]interface{})
		result["id"] = *v.ID
		result["internal_port"] = *v.InternalPort
		result["external_ip"] = *v.ExternalPort
		result["external_port"] = *v.ExternalIP

		results = append(results, result)
	}

	return results
}

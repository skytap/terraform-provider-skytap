package skytap

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func flattenNetworkInterfaces(interfaces []skytap.Interface, externalIPMaps []map[string]interface{}) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range interfaces {
		results = append(results, flattenNetworkInterface(v, externalIPMaps))
	}

	return schema.NewSet(networkInterfaceHash, results)
}

func flattenNetworkInterface(v skytap.Interface,
	externalIPMaps []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	result["interface_type"] = string(*v.NICType)
	result["network_id"] = *v.NetworkID
	result["ip"] = *v.IP
	result["hostname"] = *v.Hostname
	if len(v.Services) > 0 {
		result["published_service"] = flattenPublishedServices(v.Services, *v.IP,
			externalIPMaps)
	}
	return result
}

func flattenPublishedServices(publishedServices []skytap.PublishedService, ip string,
	externalIPMaps []map[string]interface{}) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range publishedServices {
		results = append(results, flattenPublishedService(v, ip, externalIPMaps))
	}

	return schema.NewSet(publishedServiceHash, results)
}

func flattenPublishedService(v skytap.PublishedService, ip string, externalIPMaps []map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = *v.ID
	result["internal_port"] = *v.InternalPort
	result["external_ip"] = *v.ExternalIP
	result["external_port"] = *v.ExternalPort
	key := strings.Replace(ip, ".", "-", -1) + "_" + strconv.Itoa(*v.InternalPort)
	externalIPMaps[0][key] = *v.ExternalIP
	externalIPMaps[1][key] = *v.ExternalPort
	return result
}

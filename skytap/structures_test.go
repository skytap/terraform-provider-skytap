package skytap

import (
	"encoding/json"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
	"testing"
)

const exampleInterfaceListResponse = `[
    {
        "id": "nic-20246343-38367563-0",
        "ip": "192.168.0.1",
        "hostname": "wins2016s",
        "mac": "00:50:56:11:7D:D9",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "network_name": "tftest-network-1",
        "network_url": "https://cloud.skytap.com/v2/configurations/40064014/networks/23917287",
        "network_type": "automatic",
        "network_subnet": "192.168.0.0/16",
        "nic_type": "vmxnet3",
        "secondary_ips": [],
        "public_ip_attachments": []
    },
    {
        "id": "nic-20246343-38367563-5",
        "ip": null,
        "hostname": null,
        "mac": "00:50:56:07:40:3F",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "nic_type": "e1000",
        "secondary_ips": [],
        "public_ip_attachments": []
    }
]`

const examplePublishedServiceListResponse = `[
    {
        "id": "8080",
        "internal_port": 8080,
        "external_ip": "services-uswest.skytap.com",
        "external_port": 26160
    },
    {
        "id": "8081",
        "internal_port": 8081,
        "external_ip": "services-uswest.skytap.com",
        "external_port": 17785
    }
]`

func TestFlattenInterfaces(t *testing.T) {
	var interfaces []skytap.Interface
	json.Unmarshal([]byte(exampleInterfaceListResponse), &interfaces)
	var hcl = make([]map[string]interface{}, 0)
	hcl = flattenInterfaces(interfaces)
	assert.True(t, len(hcl) == 2)
}

func TestFlattenPublishedServices(t *testing.T) {
	var publishedServices []skytap.PublishedService
	json.Unmarshal([]byte(examplePublishedServiceListResponse), &publishedServices)
	var hcl = make([]map[string]interface{}, 0)
	hcl = flattenPublishedServices(publishedServices)
	assert.True(t, len(hcl) == 2)
}

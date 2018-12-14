package skytap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
)

func TestFlattenInterfaces(t *testing.T) {

	response := string(readTestFile(t, "vm_interface_response.json"))

	expected := make(map[string][]string)
	expected["ip"] = []string{"192.168.0.1", "192.168.0.2"}
	expected["hostname"] = []string{"wins2016s", "wins2016s2"}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var interfaces []skytap.Interface
	err := json.Unmarshal([]byte(response), &interfaces)
	if err != nil {
		t.Fatal(err)
	}
	var networkInterfaces = make([]map[string]interface{}, 0)
	for _, v := range interfaces {
		networkInterfaces = append(networkInterfaces, flattenNetworkInterface(v, ipMaps))
	}
	assert.True(t, len(networkInterfaces) == 2, fmt.Sprintf("expecting: %d but received: %d", 2, len(networkInterfaces)))
	for i := 0; i < len(networkInterfaces); i++ {
		networkInterface := networkInterfaces[i]
		for key, value := range expected {
			assert.Equal(t, value[i], networkInterface[key],
				fmt.Sprintf("expecting: %s but received: %s", value[i], networkInterface[key]))
		}
	}
}

func TestFlattenPublishedServices(t *testing.T) {

	response := string(readTestFile(t, "vm_interface_services_response.json"))

	expected := make(map[string][]string)
	expected["id"] = []string{"8080", "8081"}
	expected["internal_port"] = []string{"8080", "8081"}
	expected["external_ip"] = []string{"services-uswest.skytap.com", "services-useast.skytap.com"}
	expected["external_port"] = []string{"26160", "17785"}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var services []skytap.PublishedService
	err := json.Unmarshal([]byte(response), &services)
	if err != nil {
		t.Fatal(err)
	}
	var publishedServices = make([]map[string]interface{}, 0)
	for _, v := range services {
		publishedServices = append(publishedServices, flattenPublishedService(v, "", ipMaps))
	}
	assert.True(t, len(publishedServices) == 2, fmt.Sprintf("expecting: %d but received: %d", 2, len(publishedServices)))
	for i := 0; i < len(publishedServices); i++ {
		publishedService := publishedServices[i]
		for key, value := range expected {
			if _, ok := publishedService[key].(string); ok {
				assert.Equal(t, value[i], publishedService[key].(string),
					fmt.Sprintf("expecting: %s but received: %s", value[i], publishedService[key]))
			} else {
				valueAsString := strconv.Itoa(publishedService[key].(int))
				assert.Equal(t, value[i], valueAsString,
					fmt.Sprintf("expecting: %s but received: %s", value[i], valueAsString))
			}
		}
	}
}

func TestFlattenInterfacesIPMap(t *testing.T) {

	response := string(readTestFile(t, "vm_interface_ports_response.json"))

	expectedKeys1 := []string{"192-168-0-1_8080", "192-168-0-1_8081", "192-168-0-2_8080", "192-168-0-2_8081"}
	expectedValues1 := []string{"services-uswest.skytap.com", "services-useast.skytap.com", "services-uswest.skytap.com", "services-useast.skytap.com"}
	expectedKeys2 := []string{"192-168-0-1_8080", "192-168-0-1_8081", "192-168-0-2_8080", "192-168-0-2_8081"}
	expectedValues2 := []int{26160, 17785, 26160, 17785}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var interfaces []skytap.Interface
	err := json.Unmarshal([]byte(response), &interfaces)
	if err != nil {
		t.Fatal(err)
	}
	var networkInterfaces = make([]map[string]interface{}, 0)
	for _, v := range interfaces {
		networkInterfaces = append(networkInterfaces, flattenNetworkInterface(v, ipMaps))
	}

	count := 0
	for _, key := range expectedKeys1 {
		assert.Equal(t, expectedValues1[count], ipMaps[0][key], "value")
		count++
	}
	count = 0
	for _, key := range expectedKeys2 {
		assert.Equal(t, expectedValues2[count], ipMaps[1][key], "value")
		count++
	}
}

func readTestFile(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

package skytap

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/skytap/terraform-provider-skytap/skytap/utils"
	"log"
	"testing"
)

const (
	networkEnvironmentPrefix = "tftest-net"
	testTemplate             = "1448141"
)

func init() {
	resource.AddTestSweepers("skytap_network", &resource.Sweeper{
		Name: "skytap_network",
		F:    testSweepSkytapNetwork,
	})
}

func testSweepSkytapNetwork(region string) error {
	meta, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	client := meta.environmentsClient
	ctx := meta.StopContext

	log.Printf("[INFO] Retrieving list of environments")
	environments, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving list of environments: %v", err)
	}

	for _, e := range environments.Value {
		if shouldSweepAcceptanceTestResourceWithPrefix(*e.Name, networkEnvironmentPrefix) {
			log.Printf("destroying environment %s", *e.Name)
			if err := client.Delete(ctx, *e.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccSkytapNetwork_Basic(t *testing.T) {
	uniqueSuffixEnv := acctest.RandInt()
	uniqueSuffixNet := acctest.RandInt()
	var network skytap.Network

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapNetworkConfig_basic(testTemplate, uniqueSuffixEnv, uniqueSuffixNet, "skytap.io", "192.168.1.0/24", "", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapNetworkExists("skytap_environment.foo", "skytap_network.bar", &network),
					resource.TestCheckResourceAttrSet("skytap_network.bar", "environment_id"),
					resource.TestCheckResourceAttr("skytap_network.bar", "name", fmt.Sprintf("tftest-network-%d", uniqueSuffixNet)),
					resource.TestCheckResourceAttr("skytap_network.bar", "domain", "skytap.io"),
					resource.TestCheckResourceAttr("skytap_network.bar", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttrSet("skytap_network.bar", "gateway"),
					resource.TestCheckResourceAttr("skytap_network.bar", "tunnelable", "true"),
				),
			},
		},
	})
}

func TestAccSkytapNetwork_Update(t *testing.T) {
	uniqueSuffixEnv := acctest.RandInt()
	uniqueSuffixInitial := acctest.RandInt()
	uniqueSuffixUpdate := acctest.RandInt()
	var network skytap.Network

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSkytapEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapNetworkConfig_basic(testTemplate, uniqueSuffixEnv, uniqueSuffixInitial, "skytap.io", "192.168.1.0/24", "gateway = \"192.168.1.1\"", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapNetworkExists("skytap_environment.foo", "skytap_network.bar", &network),
					resource.TestCheckResourceAttr("skytap_network.bar", "name", fmt.Sprintf("tftest-network-%d", uniqueSuffixInitial)),
					resource.TestCheckResourceAttr("skytap_network.bar", "domain", "skytap.io"),
					resource.TestCheckResourceAttr("skytap_network.bar", "subnet", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("skytap_network.bar", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("skytap_network.bar", "tunnelable", "true"),
				),
			},
			{
				Config: testAccSkytapNetworkConfig_basic(testTemplate, uniqueSuffixEnv, uniqueSuffixUpdate, "skytap.com", "192.168.2.0/24", "gateway = \"192.168.2.1\"", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapNetworkExists("skytap_environment.foo", "skytap_network.bar", &network),
					resource.TestCheckResourceAttr("skytap_network.bar", "name", fmt.Sprintf("tftest-network-%d", uniqueSuffixUpdate)),
					resource.TestCheckResourceAttr("skytap_network.bar", "domain", "skytap.com"),
					resource.TestCheckResourceAttr("skytap_network.bar", "subnet", "192.168.2.0/24"),
					resource.TestCheckResourceAttr("skytap_network.bar", "gateway", "192.168.2.1"),
					resource.TestCheckResourceAttr("skytap_network.bar", "tunnelable", "false"),
				),
			},
		},
	})
}

func testAccCheckSkytapNetworkExists(environmentName string, networkName string, network *skytap.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rsEnvironment, err := getResource(s, environmentName)
		if err != nil {
			return err
		}
		environmentID := rsEnvironment.Primary.ID

		rsNetwork, err := getResource(s, networkName)
		if err != nil {
			return err
		}

		// Get the network
		net, err := getNetwork(rsNetwork, environmentID)
		if err != nil {
			return err
		}

		*network = *net
		log.Printf("[DEBUG] network (%s)\n", *network.ID)

		return nil
	}
}

func testAccSkytapNetworkConfig_basic(templateID string, uniqueSuffixEnv int, uniqueSuffixNet int, domain string, subnet string, gateway string, tunnelable bool) string {
	return fmt.Sprintf(`
	resource "skytap_environment" "foo" {
		template_id = %s
		name 		= "%s-environment-%d"
		description = "This is an environment to support a network skytap terraform provider acceptance test"
	}

	resource "skytap_network" "bar" {
  		"name"        		= "tftest-network-%d"
		"domain"      		= %q
  		"environment_id" 	= "${skytap_environment.foo.id}"
  		"subnet"      		= %q
		%s
  		"tunnelable"  		= %t
	}
`, templateID, networkEnvironmentPrefix, uniqueSuffixEnv, uniqueSuffixNet, domain, subnet, gateway, tunnelable)
}

func getNetwork(rs *terraform.ResourceState, environmentID string) (*skytap.Network, error) {
	var err error
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*SkytapClient).networksClient
	ctx := testAccProvider.Meta().(*SkytapClient).StopContext

	// Retrieve our network by referencing it's state ID for API lookup
	network, errClient := client.Get(ctx, environmentID, rs.Primary.ID)
	if errClient != nil {
		if utils.ResponseErrorIsNotFound(err) {
			err = errors.Errorf("network (%s) was not found - does not exist", rs.Primary.ID)
		}

		err = fmt.Errorf("error retrieving network (%s): %v", rs.Primary.ID, err)
	}
	return network, err
}

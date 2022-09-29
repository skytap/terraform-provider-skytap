package skytap

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/skytap/skytap-sdk-go/skytap"

	"github.com/terraform-providers/terraform-provider-skytap/skytap/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSkytapLabelICNR_Basic(t *testing.T) {
	templateID := utils.GetEnv("SKYTAP_TEMPLATE_ID", "1478959")
	uniqueSuffix := acctest.RandInt()
	var tunnel skytap.ICNRTunnel

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSkytapLabelCategoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSkytapICNRTunnel_basic("tftest", uniqueSuffix, templateID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSkytapICNRTunnelExists("skytap_icnr_tunnel.tunnel", &tunnel),
				),
			},
		},
	})
}

func testAccSkytapICNRTunnel_basic(prefix string, suffix int, templateId string) string {
	return fmt.Sprintf(`
		resource "skytap_environment" "env1" {
	    	template_id = "%s"
	    	name = "%s-environment-%d"
	    	description = "This is an environment created by the skytap terraform provider acceptance test"
      	}

		resource "skytap_environment" "env2" {
	    	template_id = "%s"
	    	name = "%s-environment-%d"
	    	description = "This is an environment created by the skytap terraform provider acceptance test"
      	}

		resource "skytap_network" "net1" {
  			environment_id = skytap_environment.env1.id
			name = "net1"
  			domain = "domain.com"
			subnet = "10.0.100.0/24"
			gateway = "10.0.100.254"
  			tunnelable = true
		}

		resource "skytap_network" "net2" {
  			environment_id = skytap_environment.env2.id
			name = "net2"
  			domain = "domain.com"
			subnet = "10.0.200.0/24"
			gateway = "10.0.200.254"
  			tunnelable = true
		}

		resource "skytap_icnr_tunnel" "tunnel" {
			source = skytap_network.net1.id
			target = skytap_network.net2.id
		}
		`,
		templateId, prefix, suffix, templateId, prefix, suffix)
}

func testAccCheckSkytapICNRTunnelExists(name string, icnrTunnel *skytap.ICNRTunnel) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		rs, err := getResource(s, name)
		if err != nil {
			return err
		}

		// retrieve the connection established in Provider configuration
		client := testAccProvider.Meta().(*SkytapClient).icnrTunnelClient
		ctx := context.TODO()

		icnrTunnel, err = client.Get(ctx, rs.Primary.ID)
		if err != nil {
			if utils.ResponseErrorIsNotFound(err) {
				err = fmt.Errorf("tunnel (%s) was not found - does not exist", rs.Primary.ID)
			}

			err = fmt.Errorf("error retrieving tunnel (%s): %v", rs.Primary.ID, err)
		}
		return
	}
}

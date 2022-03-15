package mira

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMiraAvailableSubnets(t *testing.T) {
	t.Skip("data source not yet implemented, remove this once you add your own code")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMiraAvailableSubnets,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.mira_available_subnet_data_source.foo", "sample_attribute", regexp.MustCompile("^ba")),
				),
			},
		},
	})
}

const testAccDataSourceMiraAvailableSubnets = `
data "mira_available_subnet_data_source" "foo" {
  sample_attribute = "bar"
}
`

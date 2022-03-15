package mira

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceMiraAllocatedSubnet(t *testing.T) {
	t.Skip("resource not yet implemented, remove this once you add your own code")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMiraAllocatedSubnet,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"mira_allocated_subnet_resource.foo", "sample_attribute", regexp.MustCompile("^ba")),
				),
			},
		},
	})
}

const testAccResourceMiraAllocatedSubnet = `
resource "mira_allocated_subnet_resource" "foo" {
  sample_attribute = "bar"
}
`

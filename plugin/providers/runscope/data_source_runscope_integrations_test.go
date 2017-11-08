package runscope

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

func TestAccDataSourceRunscopeIntegrations_Basic(t *testing.T) {

	teamId := os.Getenv("RUNSCOPE_TEAM_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDataSourceRunscopeIntegrationsConfig, teamId),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceRunscopeIntegrations("data.runscope_integrations.by_type"),
				),
			},
		},
	})
}

func testAccDataSourceRunscopeIntegrations(dataSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[dataSource]
		a := r.Primary.Attributes

		if len(a["integrations.#"]) == 0 {
			return fmt.Errorf("Expected to get integrations returned from runscope data resource")
		}

		if a["integrations.0.id"] == "" {
			return fmt.Errorf("Expected to get an integration ID from runscope data resource")
		}

		if a["integrations.0.description"] == "" {
			return fmt.Errorf("Expected to get an integration description from runscope data resource")
		}

		if a["integrations.0.integration_type"] == "" {
			return fmt.Errorf("Expected to get an lintegration type from runscope data resource")
		}

		return nil
	}
}

const testAccDataSourceRunscopeIntegrationsConfig = `
data "runscope_integrations" "by_type" {
	team_uuid = "%s"
	filter = {
		name = "type"
		values = ["slack"]
	}
}
`

func TestAccDataSourceRunscopeIntegrations_Filter(t *testing.T) {

	teamId := os.Getenv("RUNSCOPE_TEAM_ID")
	integrationDesc := os.Getenv("RUNSCOPE_INTEGRATION_DESC")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDataSourceRunscopeIntegrationsFilterConfig, teamId, integrationDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceRunscopeIntegrationsFilter("data.runscope_integrations.by_description"),
				),
			},
		},
	})
}

func testAccDataSourceRunscopeIntegrationsFilter(dataSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[dataSource]
		a := r.Primary.Attributes
		integrationDesc := os.Getenv("RUNSCOPE_INTEGRATION_DESC")

		if a["integrations.0.description"] != integrationDesc {
			return fmt.Errorf("Expected integration description %s to be %s", a["integrations.0.description"], integrationDesc)
		}

		return nil
	}
}

const testAccDataSourceRunscopeIntegrationsFilterConfig = `
	data "runscope_integrations" "by_description" {
		team_uuid = "%s"
		filter = {
			name = "type"
			values = ["slack"]
		}
		filter = {
			name = "description"
			values = ["%s","other test description"]
		}
	}
	`

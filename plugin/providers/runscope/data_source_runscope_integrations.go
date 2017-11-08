package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

func dataSourceRunscopeIntegrations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRunscopeIntegrationsRead,

		Schema: map[string]*schema.Schema{
			"team_uuid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"integrations": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"integration_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func dataSourceRunscopeIntegrationsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	log.Printf("[INFO] Reading Runscope integration")

	filters, filtersOk := d.GetOk("filter")

	resp, err := client.ListIntegrations(d.Get("team_uuid").(string))
	if err != nil {
		return err
	}

	var integrations []map[string]string
	for _, integration := range resp {
		if filtersOk {
			if !integrationFiltersTest(integration, filters.(*schema.Set)) {
				continue
			}
		}

		found := make(map[string]string)
		found["id"] = integration.ID
		found["description"] = integration.Description
		found["integration_type"] = integration.IntegrationType
		integrations = append(integrations, found)
	}

	if len(integrations) == 0 {
		return fmt.Errorf("Unable to locate any integrations")
	}

	d.SetId(time.Now().UTC().String())
	d.Set("integrations", integrations)

	return nil
}

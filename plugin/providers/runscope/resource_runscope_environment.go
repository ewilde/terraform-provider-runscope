package runscope

import (
	"fmt"
	"github.com/ewilde/go-runscope"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceRunscopeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceEnvironmentCreate,
		Read:   resourceEnvironmentRead,
		Update: resourceEnvironmentUpdate,
		Delete: resourceEnvironmentDelete,

		Schema: map[string]*schema.Schema{
			"bucket_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"test_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"script": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},
			"preserve_cookies": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
			},
			"initial_variables": &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				ForceNew: false,
			},
			"integrations": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"regions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"remote_agents": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"uuid": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Optional: true,
			},
			"retry_on_failure": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	name := d.Get("name").(string)
	log.Printf("[INFO] Creating environment with name: %s", name)

	environment, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] environment create: %#v", environment)

	var createdEnvironment *runscope.Environment
	bucketId := d.Get("bucket_id").(string)

	if testId, ok := d.GetOk("test_id"); ok {
		createdEnvironment, err = client.CreateTestEnvironment(environment,
			&runscope.Test{ID: testId.(string), Bucket: &runscope.Bucket{Key: bucketId}})
	} else {
		createdEnvironment, err = client.CreateSharedEnvironment(environment,
			&runscope.Bucket{Key: bucketId})
	}
	if err != nil {
		return fmt.Errorf("Failed to create environment: %s", err)
	}

	d.SetId(createdEnvironment.ID)
	log.Printf("[INFO] environment ID: %s", d.Id())

	return resourceEnvironmentRead(d, meta)
}

func resourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	environmentFromResource, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Failed to read environment from resource data: %s", err)
	}

	var environment *runscope.Environment
	bucketId := d.Get("bucket_id").(string)
	if testId, ok := d.GetOk("test_id"); ok {
		environment, err = client.ReadTestEnvironment(
			environmentFromResource, &runscope.Test{ID: testId.(string), Bucket: &runscope.Bucket{Key: bucketId}})
	} else {
		environment, err = client.ReadSharedEnvironment(
			environmentFromResource, &runscope.Bucket{Key: bucketId})
	}

	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "403") {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Couldn't find environment: %s", err)
	}

	d.Set("bucket_id", bucketId)
	d.Set("test_id", d.Get("test_id").(string))
	d.Set("name", environment.Name)
	d.Set("script", environment.Script)
	d.Set("preserve_cookies", environment.PreserveCookies)
	d.Set("initial_variables", environment.InitialVariables)
	d.Set("integrations", readIntegrations(environment.Integrations))
	d.Set("retry_on_failure", environment.RetryOnFailure)
	return nil
}

func resourceEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(false)
	environment, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Error updating environment: %s", err)
	}

	if d.HasChange("name") ||
		d.HasChange("script") ||
		d.HasChange("preserve_cookies") ||
		d.HasChange("initial_variables") ||
		d.HasChange("integrations") ||
		d.HasChange("regions") ||
		d.HasChange("remote_agents") ||
		d.HasChange("retry_on_failure") {
		client := meta.(*runscope.Client)
		bucketId := d.Get("bucket_id").(string)
		if testId, ok := d.GetOk("test_id"); ok {
			_, err = client.UpdateTestEnvironment(
				environment, &runscope.Test{ID: testId.(string), Bucket: &runscope.Bucket{Key: bucketId}})
		} else {
			_, err = client.UpdateSharedEnvironment(
				environment, &runscope.Bucket{Key: bucketId})
		}
		if err != nil {
			return fmt.Errorf("Error updating environment: %s", err)
		}
	}

	return nil
}

func resourceEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*runscope.Client)

	environmentFromResource, err := createEnvironmentFromResourceData(d)
	if err != nil {
		return fmt.Errorf("Failed to read environment from resource data: %s", err)
	}

	bucketId := d.Get("bucket_id").(string)
	if testId, ok := d.GetOk("test_id"); ok {
		log.Printf("[INFO] Deleting test environment with id: %s name: %s, from test %s",
			environmentFromResource.ID, environmentFromResource.Name, testId.(string))
		err = client.DeleteEnvironment(
			environmentFromResource, &runscope.Bucket{Key: bucketId})
	} else {
		log.Printf("[INFO] Deleting shared environment with id: %s name: %s",
			environmentFromResource.ID, environmentFromResource.Name)
		err = client.DeleteEnvironment(
			environmentFromResource, &runscope.Bucket{Key: bucketId})
	}

	if err != nil {
		return fmt.Errorf("Error deleting environment: %s", err)
	}

	return nil
}

func createEnvironmentFromResourceData(d *schema.ResourceData) (*runscope.Environment, error) {

	environment := runscope.NewEnvironment()
	environment.ID = d.Id()

	if attr, ok := d.GetOk("name"); ok {
		environment.Name = attr.(string)
	}

	if attr, ok := d.GetOk("test_id"); ok {
		environment.TestID = attr.(string)
	}

	if attr, ok := d.GetOk("script"); ok {
		environment.Script = attr.(string)
	}

	if attr, ok := d.GetOk("preserve_cookies"); ok {
		environment.PreserveCookies = attr.(bool)
	}

	if attr, ok := d.GetOk("initial_variables"); ok {
		variablesRaw := attr.(map[string]interface{})
		variables := map[string]string{}
		for k, v := range variablesRaw {
			variables[k] = v.(string)
		}

		environment.InitialVariables = variables
	}

	if attr, ok := d.GetOk("integrations"); ok {
		integrations := []*runscope.EnvironmentIntegration{}
		items := attr.([]interface{})
		for _, item := range items {
			integration := runscope.EnvironmentIntegration{
				ID: item.(string),
			}

			integrations = append(integrations, &integration)
		}

		environment.Integrations = integrations
	}

	if attr, ok := d.GetOk("regions"); ok {
		regions := []string{}
		items := attr.([]interface{})
		for _, x := range items {
			item := x.(string)
			regions = append(regions, item)
		}

		environment.Regions = regions
	}

	if attr, ok := d.GetOk("remote_agents"); ok {
		remote_agents := []*runscope.LocalMachine{}
		items := attr.([]interface{})
		for _, x := range items {
			item := x.(map[string]interface{})
			remote_agent := runscope.LocalMachine{
				Name: item["name"].(string),
				UUID: item["uuid"].(string),
			}

			remote_agents = append(remote_agents, &remote_agent)
		}

		environment.RemoteAgents = remote_agents
	}

	if attr, ok := d.GetOk("retry_on_failure"); ok {
		environment.RetryOnFailure = attr.(bool)
	}

	return environment, nil
}

func readIntegrations(integrations []*runscope.EnvironmentIntegration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(integrations))
	for _, integration := range integrations {

		item := map[string]interface{}{
			"id":               integration.ID,
			"integration_type": integration.IntegrationType,
			"description":      integration.Description,
		}

		result = append(result, item)
	}

	return result
}

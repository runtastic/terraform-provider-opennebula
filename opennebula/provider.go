package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to your public or private OpenNebula",
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user to identify as",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The password for the user",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"opennebula_template": resourceTemplate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Endpoint: d.Get("endpoint").(string),
		User: d.Get("user").(string),
		Password: d.Get("password").(string),
	}
	return config.Client()
}

package circleci

import (
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CIRCLECI_API_TOKEN", nil),
				Description: "Token to use to authenticate to CircleCI.",
			},
		},

		ConfigureFunc: providerConfigure,

		ResourcesMap: map[string]*schema.Resource{
			"circleci_project": resourceProject(),
		},
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := &ApiClient{
		Token:      d.Get("api_token").(string),
		HTTPClient: cleanhttp.DefaultClient(),
		Debug:      true,
	}

	client.HTTPClient.Transport = logging.NewTransport("CircleCI", client.HTTPClient.Transport)

	return client, nil
}

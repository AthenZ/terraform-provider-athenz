package athenz

import (
	"fmt"
	"os"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"zms_url": {
				Type:        schema.TypeString,
				Description: fmt.Sprintf("Athenz API URL"),
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_ZMS_URL", nil),
			},
			"cert": {
				Type:        schema.TypeString,
				Description: fmt.Sprintf("Athenz client certificate"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_CERT", os.Getenv("HOME")+"/.athenz/cert"),
			},
			"key": {
				Type:        schema.TypeString,
				Description: fmt.Sprintf("Athenz client key"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_KEY", os.Getenv("HOME")+"/.athenz/key"),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"athenz_role":               DataSourceRole(),
			"athenz_group":              DataSourceGroup(),
			"athenz_policy":             DataSourcePolicy(),
			"athenz_policy_version":     DataSourcePolicyVersion(),
			"athenz_service":            dataSourceService(),
			"athenz_domain":             DataSourceDomain(),
			"athenz_all_domain_details": DataSourceAllDomainDetails(),
			"athenz_roles":              DataSourceRoles(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"athenz_role":             ResourceRole(),
			"athenz_group":            ResourceGroup(),
			"athenz_policy":           ResourcePolicy(),
			"athenz_policy_version":   ResourcePolicyVersion(),
			"athenz_service":          ResourceService(),
			"athenz_sub_domain":       ResourceSubDomain(),
			"athenz_user_domain":      ResourceUserDomain(),
			"athenz_top_level_domain": ResourceTopLevelDomain(),
		},

		ConfigureFunc: configProvider,
	}
}

func configProvider(d *schema.ResourceData) (interface{}, error) {
	zms := client.ZmsConfig{
		Url:  d.Get("zms_url").(string),
		Cert: d.Get("cert").(string),
		Key:  d.Get("key").(string),
	}
	if zms.Url == "localhost" {
		return client.AccTestZmsClient()
	}
	return client.NewClient(zms.Url, zms.Cert, zms.Key)
}

package athenz

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
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
			"cacert": {
				Type:        schema.TypeString,
				Description: fmt.Sprintf("CA Certificate file path"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_CA_CERT", ""),
			},
			"disable_resource_ownership": {
				Type:        schema.TypeBool,
				Description: fmt.Sprintf("Disable resource ownership feature"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_DISABLE_RESOURCE_OWNERSHIP", false),
			},
			"resource_owner": {
				Type:        schema.TypeString,
				Description: fmt.Sprintf("Resource Owner Identity"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_RESOURCE_OWNER", "TF"),
			},
			"role_meta_resource_state": {
				Type:        schema.TypeInt,
				Description: fmt.Sprintf("Default state for athenz_role_meta resources"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_ROLE_META_RESOURCE_STATE", client.StateCreateIfNecessary),
			},
			"group_meta_resource_state": {
				Type:        schema.TypeInt,
				Description: fmt.Sprintf("Default state for athenz_group_meta resources"),
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENZ_GROUP_META_RESOURCE_STATE", client.StateCreateIfNecessary),
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
			"athenz_role":                     ResourceRole(),
			"athenz_role_members":             ResourceRoleMembers(),
			"athenz_self_serve_role_members":  ResourceSelfServeRoleMembers(),
			"athenz_role_meta":                ResourceRoleMeta(),
			"athenz_group":                    ResourceGroup(),
			"athenz_group_members":            ResourceGroupMembers(),
			"athenz_self_serve_group_members": ResourceSelfServeGroupMembers(),
			"athenz_group_meta":               ResourceGroupMeta(),
			"athenz_policy":                   ResourcePolicy(),
			"athenz_policy_version":           ResourcePolicyVersion(),
			"athenz_service":                  ResourceService(),
			"athenz_sub_domain":               ResourceSubDomain(),
			"athenz_user_domain":              ResourceUserDomain(),
			"athenz_top_level_domain":         ResourceTopLevelDomain(),
			"athenz_domain_meta":              ResourceDomainMeta(),
		},

		ConfigureContextFunc: configProvider,
	}
}

func configProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	zms := client.ZmsConfig{
		Url:                    d.Get("zms_url").(string),
		Cert:                   d.Get("cert").(string),
		Key:                    d.Get("key").(string),
		CaCert:                 d.Get("cacert").(string),
		RoleMetaResourceState:  d.Get("role_meta_resource_state").(int),
		GroupMetaResourceState: d.Get("group_meta_resource_state").(int),
	}
	// if resource ownership is not disabled, then load the resource owner
	if !d.Get("disable_resource_ownership").(bool) {
		zms.ResourceOwner = d.Get("resource_owner").(string)
	}
	c, err := client.NewClient(&zms)
	return c, diag.FromErr(err)
}

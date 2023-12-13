package athenz

import (
	"context"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/AthenZ/terraform-provider-athenz/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDomainMeta() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDomainMetaCreate,
		ReadContext:   resourceDomainMetaRead,
		UpdateContext: resourceDomainMetaUpdate,
		DeleteContext: resourceDomainMetaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "name of the domain",
				Required:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "description for the domain",
				Optional:    true,
			},
			"application_id": {
				Type:        schema.TypeString,
				Description: "associated application id",
				Optional:    true,
			},
			"user_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "all user members in the domain will have specified max expiry days",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"token_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "tokens issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"service_cert_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "service identity certs issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"role_cert_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "role certs issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"service_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "all services in the domain roles will have specified max expiry days",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"group_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "all groups in the domain roles will have specified max expiry days",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"member_purge_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "purge role/group members with expiry date configured days in the past",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"user_authority_filter": {
				Type:        schema.TypeString,
				Description: "membership filtered based on user authority configured attributes",
				Optional:    true,
			},
			"business_service": {
				Type:        schema.TypeString,
				Description: "associated business service with domain",
				Optional:    true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"contacts": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"audit_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  AUDIT_REF,
			},
		},
	}
}

func resourceDomainMetaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)

	dn := d.Get("domain").(string)
	resp := updateDomainMeta(zmsClient, dn, d)
	if resp != nil {
		return resp
	}
	d.SetId(dn)
	return readAfterWrite(resourceDomainMetaRead, ctx, d, meta)
}

func resourceDomainMetaRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	domain, err := zmsClient.GetDomain(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("domain", domain.Name); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("application_id", domain.ApplicationId); err != nil {
		return diag.FromErr(err)
	}
	if domain.TokenExpiryMins != nil {
		if err = d.Set("token_expiry_mins", domain.TokenExpiryMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if domain.ServiceCertExpiryMins != nil {
		if err = d.Set("service_cert_expiry_mins", domain.ServiceCertExpiryMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if domain.RoleCertExpiryMins != nil {
		if err = d.Set("role_cert_expiry_mins", domain.RoleCertExpiryMins); err != nil {
			return diag.FromErr(err)
		}
	}
	if domain.ServiceExpiryDays != nil {
		if err = d.Set("service_expiry_days", domain.ServiceExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if domain.GroupExpiryDays != nil {
		if err = d.Set("group_expiry_days", domain.GroupExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if domain.MemberPurgeExpiryDays != nil {
		if err = d.Set("member_purge_expiry_days", domain.MemberPurgeExpiryDays); err != nil {
			return diag.FromErr(err)
		}
	}
	if err = d.Set("user_authority_filter", domain.UserAuthorityFilter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("business_service", domain.BusinessService); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("tags", flattenTag(domain.Tags)); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("contacts", domain.Contacts); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDomainMetaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	resp := updateDomainMeta(zmsClient, d.Id(), d)
	if resp != nil {
		return resp
	}
	return readAfterWrite(resourceDomainMetaRead, ctx, d, meta)
}

func resourceDomainMetaDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	auditRef := d.Get("audit_ref").(string)
	var zero int32
	zero = 0
	domainMeta := zms.DomainMeta{
		Description:           "",
		ApplicationId:         "",
		MemberExpiryDays:      &zero,
		TokenExpiryMins:       &zero,
		ServiceCertExpiryMins: &zero,
		RoleCertExpiryMins:    &zero,
		ServiceExpiryDays:     &zero,
		GroupExpiryDays:       &zero,
		MemberPurgeExpiryDays: &zero,
		UserAuthorityFilter:   "",
		BusinessService:       "",
		Tags:                  make(map[zms.TagKey]*zms.TagValueList),
		Contacts:              make(map[zms.SimpleName]string),
	}
	if v, ok := d.GetOk("tags"); ok {
		for key := range v.(map[string]interface{}) {
			domainMeta.Tags[zms.TagKey(key)] = &zms.TagValueList{List: []zms.TagCompoundValue{}}
		}
	}
	if v, ok := d.GetOk("contacts"); ok {
		for key := range v.(map[string]interface{}) {
			domainMeta.Contacts[zms.SimpleName(key)] = ""
		}
	}
	err := zmsClient.PutDomainMeta(d.Id(), auditRef, &domainMeta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateDomainMeta(zmsClient client.ZmsClient, dn string, d *schema.ResourceData) diag.Diagnostics {

	domain, err := zmsClient.GetDomain(dn)
	if err != nil {
		return diag.Errorf("domain %s does not exist", dn)
	}
	domainMeta := zms.DomainMeta{
		Description:           domain.Description,
		ApplicationId:         domain.ApplicationId,
		MemberExpiryDays:      domain.MemberExpiryDays,
		TokenExpiryMins:       domain.TokenExpiryMins,
		ServiceCertExpiryMins: domain.ServiceCertExpiryMins,
		RoleCertExpiryMins:    domain.RoleCertExpiryMins,
		ServiceExpiryDays:     domain.ServiceExpiryDays,
		GroupExpiryDays:       domain.GroupExpiryDays,
		MemberPurgeExpiryDays: domain.MemberPurgeExpiryDays,
		UserAuthorityFilter:   domain.UserAuthorityFilter,
		BusinessService:       domain.BusinessService,
		Tags:                  domain.Tags,
		Contacts:              domain.Contacts,
	}
	domainMeta.Description = d.Get("description").(string)
	domainMeta.ApplicationId = d.Get("application_id").(string)
	if v, ok := d.GetOk("user_expiry_days"); ok {
		memberExpiryDays := int32(v.(int))
		domainMeta.MemberExpiryDays = &memberExpiryDays
	}
	if v, ok := d.GetOk("token_expiry_mins"); ok {
		tokenExpiryMins := int32(v.(int))
		domainMeta.TokenExpiryMins = &tokenExpiryMins
	}
	if v, ok := d.GetOk("service_cert_expiry_mins"); ok {
		serviceCertExpiryMins := int32(v.(int))
		domainMeta.ServiceCertExpiryMins = &serviceCertExpiryMins
	}
	if v, ok := d.GetOk("role_cert_expiry_mins"); ok {
		roleCertExpiryMins := int32(v.(int))
		domainMeta.RoleCertExpiryMins = &roleCertExpiryMins
	}
	if v, ok := d.GetOk("service_expiry_days"); ok {
		serviceExpiryDays := int32(v.(int))
		domainMeta.ServiceExpiryDays = &serviceExpiryDays
	}
	if v, ok := d.GetOk("group_expiry_days"); ok {
		groupExpiryDays := int32(v.(int))
		domainMeta.GroupExpiryDays = &groupExpiryDays
	}
	if v, ok := d.GetOk("member_purge_expiry_days"); ok {
		memberPurgeExpiryDays := int32(v.(int))
		domainMeta.MemberPurgeExpiryDays = &memberPurgeExpiryDays
	}
	domainMeta.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	domainMeta.BusinessService = d.Get("business_service").(string)
	if v, ok := d.GetOk("tags"); ok {
		domainMeta.Tags = expandTagsMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("contacts"); ok {
		domainMeta.Contacts = expandContactsMap(v.(map[string]interface{}))
	}
	auditRef := d.Get("audit_ref").(string)
	err = zmsClient.PutDomainMeta(dn, auditRef, &domainMeta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandContactsMap(contactsMap map[string]interface{}) map[zms.SimpleName]string {
	contacts := map[zms.SimpleName]string{}
	for key, val := range contactsMap {
		contacts[zms.SimpleName(key)] = val.(string)
	}
	return contacts
}

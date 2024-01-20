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
				ValidateFunc: validation.IntAtLeast(0),
			},
			"token_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "tokens issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"service_cert_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "service identity certs issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"role_cert_expiry_mins": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "role certs issued for this domain will have specified max timeout in mins",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"service_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "all services in the domain roles will have specified max expiry days",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"group_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "all groups in the domain roles will have specified max expiry days",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"member_purge_expiry_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "purge role/group members with expiry date configured days in the past",
				ValidateFunc: validation.IntAtLeast(0),
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
	if domain.MemberExpiryDays != nil {
		if err = d.Set("user_expiry_days", domain.MemberExpiryDays); err != nil {
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
	domainMeta.UserAuthorityFilter = d.Get("user_authority_filter").(string)
	domainMeta.BusinessService = d.Get("business_service").(string)
	if d.HasChange("user_expiry_days") {
		memberExpiryDays := int32(d.Get("user_expiry_days").(int))
		domainMeta.MemberExpiryDays = &memberExpiryDays
	}
	if d.HasChange("token_expiry_mins") {
		tokenExpiryMins := int32(d.Get("token_expiry_mins").(int))
		domainMeta.TokenExpiryMins = &tokenExpiryMins
	}
	if d.HasChange("service_cert_expiry_mins") {
		serviceCertExpiryMins := int32(d.Get("service_cert_expiry_mins").(int))
		domainMeta.ServiceCertExpiryMins = &serviceCertExpiryMins
	}
	if d.HasChange("role_cert_expiry_mins") {
		roleCertExpiryMins := int32(d.Get("role_cert_expiry_mins").(int))
		domainMeta.RoleCertExpiryMins = &roleCertExpiryMins
	}
	if d.HasChange("service_expiry_days") {
		serviceExpiryDays := int32(d.Get("service_expiry_days").(int))
		domainMeta.ServiceExpiryDays = &serviceExpiryDays
	}
	if d.HasChange("group_expiry_days") {
		groupExpiryDays := int32(d.Get("group_expiry_days").(int))
		domainMeta.GroupExpiryDays = &groupExpiryDays
	}
	if d.HasChange("member_purge_expiry_days") {
		memberPurgeExpiryDays := int32(d.Get("member_purge_expiry_days").(int))
		domainMeta.MemberPurgeExpiryDays = &memberPurgeExpiryDays
	}
	if d.HasChange("tags") {
		_, n := d.GetChange("tags")
		domainMeta.Tags = expandTagsMap(n.(map[string]interface{}))
	}
	if d.HasChange("contacts") {
		_, n := d.GetChange("contacts")
		domainMeta.Contacts = expandContactsMap(n.(map[string]interface{}))
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

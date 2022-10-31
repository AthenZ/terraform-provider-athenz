package athenz

import (
	"context"
	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/AthenZ/terraform-provider-athenz/client"
	"github.com/ardielle/ardielle-go/rdl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func ResourcePolicyVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourcePolicyVersionRead,
		CreateContext: resourcePolicyVersionCreate,
		UpdateContext: resourcePolicyVersionUpdate,
		DeleteContext: resourcePolicyVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:             schema.TypeString,
				Description:      "Name of the domain that policy belongs to",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(DOMAIN_NAME),
			},
			"name": {
				Type:             schema.TypeString,
				Description:      "Name of the policy",
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validatePatternFunc(ENTTITY_NAME),
			},
			"active_version": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The policy version that will be active",
				ValidateDiagFunc: validatePatternFunc(SIMPLE_NAME),
			},
			"version": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validatePatternFunc(SIMPLE_NAME),
						},
						"assertion": resourceAssertionSchema(),
					},
					// utilized CustomizeDiff method to achieve multi-attribute validation at terraform plan stage
					CustomizeDiff: validateAssertion(),
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

func resourcePolicyVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain", dn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", pn); err != nil {
		return diag.FromErr(err)
	}
	policyVersionList, err := getAllPolicyVersions(zmsClient, dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			log.Printf("[WARN] Athenz Policy %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving Athenz Policy %s: %s", d.Id(), v)
	case rdl.Any:
		return diag.FromErr(err)
	}

	if policyVersionList == nil {
		return diag.Errorf("error retrieving Athenz Policy - Make sure your cert/key are valid")
	}

	activeVersion := getActiveVersionName(policyVersionList)
	if activeVersion == "" {
		return diag.Errorf("not found active version for the policy: %s", d.Id())
	}
	if err = d.Set("active_version", activeVersion); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("version", flattenPolicyVersions(policyVersionList)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourcePolicyVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn := d.Get("domain").(string)
	pn := d.Get("name").(string)
	fullResourceName := dn + POLICY_SEPARATOR + pn
	policyCheck, err := zmsClient.GetPolicy(dn, pn)
	switch v := err.(type) {
	case rdl.ResourceError:
		if v.Code == 404 {
			auditRef := d.Get("audit_ref").(string)
			activeVersion := d.Get("active_version").(string)
			versions := d.Get("version").(*schema.Set).List()
			if err := validateSchema(activeVersion, versions); err != nil {
				return diag.FromErr(err)
			}
			policyVersions := make([]zms.Policy, 0, len(versions))
			var activeVersionIndex int
			for i, version := range versions {
				versionName, versionAssertions := expandPolicyVersion(version, dn)
				active := versionName == activeVersion
				if active {
					activeVersionIndex = i
				}
				policyVersion := zms.Policy{
					Name:       zms.ResourceName(fullResourceName),
					Version:    zms.SimpleName(versionName),
					Active:     &active,
					Assertions: versionAssertions,
				}
				policyVersions = append(policyVersions, policyVersion)
			}
			// must put the active version first
			policyVersions[0], policyVersions[activeVersionIndex] = policyVersions[activeVersionIndex], policyVersions[0]
			for _, policyVersion := range policyVersions {
				if err := zmsClient.PutPolicy(dn, pn, auditRef, &policyVersion); err != nil {
					return diag.FromErr(err)
				}
			}
		} else {
			return diag.FromErr(err)
		}
	case rdl.Any:
		return diag.FromErr(err)
	case nil:
		if policyCheck != nil {
			return diag.Errorf("the policy %s is already exists in the domain %s use terraform import command", pn, dn)
		} else {
			return diag.FromErr(err)
		}
	}
	d.SetId(fullResourceName)
	return resourcePolicyVersionRead(ctx, d, meta)
}

func validateSchema(activeVersion string, versions []interface{}) error {
	versionNameList := make([]string, 0, len(versions))
	for _, version := range versions {
		versionName := version.(map[string]interface{})["version_name"].(string)
		versionNameList = append(versionNameList, versionName)
	}
	err := validateVersionNameList(versionNameList)
	if err != nil {
		return err
	}
	return validateActiveVersion(activeVersion, versionNameList)
}
func resourcePolicyVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	policyVersionList, err := getAllPolicyVersions(zmsClient, dn, pn)
	if err != nil {
		return diag.Errorf("error retrieving Athenz Policy vrsions: %s", err)
	}
	activeVersion := d.Get("active_version").(string)
	versions := d.Get("version").(*schema.Set).List()
	auditRef := d.Get("audit_ref").(string)
	if err = validateSchema(activeVersion, versions); err != nil {
		return diag.FromErr(err)
	}
	if d.HasChange("version") {
		oldVersions, newVersions := handleChange(d, "version")
		versionsToPut := newVersions.Difference(oldVersions).List()
		versionsToDelete := getVersionsNamesToRemove(oldVersions.Difference(newVersions).List(), versionsToPut)
		for _, version := range versionsToPut {
			policyVersion := version.(map[string]interface{})
			versionName := policyVersion["version_name"].(string)
			if versionName != "" { // understand why this is happening during the update
				zmsPolicyVersion := findPolicyVersion(policyVersionList, versionName)
				if zmsPolicyVersion == nil {
					zmsPolicyVersion = zms.NewPolicy()
					zmsPolicyVersion.Name = zms.ResourceName(dn + POLICY_SEPARATOR + pn)
					zmsPolicyVersion.Version = zms.SimpleName(versionName)
					// at first, each new version is added as inactive
					active := false
					zmsPolicyVersion.Active = &active
				}
				assertions := expandPolicyAssertions(dn, policyVersion["assertion"].(*schema.Set).List())
				zmsPolicyVersion.Assertions = assertions
				if err = zmsClient.PutPolicy(dn, pn, auditRef, zmsPolicyVersion); err != nil {
					return diag.FromErr(err)
				}
			}
		}
		// check, after all new versions have been added
		if d.HasChange("active_version") {
			policyOptions := zms.PolicyOptions{
				Version: zms.SimpleName(activeVersion),
			}
			if err = zmsClient.SetActivePolicyVersion(dn, pn, &policyOptions, auditRef); err != nil {
				return diag.FromErr(err)
			}
		}
		for _, versionName := range versionsToDelete {
			if err = zmsClient.DeletePolicyVersion(dn, pn, versionName, auditRef); err != nil {
				return diag.Errorf("can't remove the policy:%s, version:%s. the error:%s", dn+POLICY_SEPARATOR+pn, versionName, err)
			}
		}
	} else if d.HasChange("active_version") {
		policyOptions := zms.PolicyOptions{
			Version: zms.SimpleName(activeVersion),
		}
		if err = zmsClient.SetActivePolicyVersion(dn, pn, &policyOptions, auditRef); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourcePolicyVersionRead(ctx, d, meta)
}

func findPolicyVersion(policyVersions []*zms.Policy, lookingVersion string) *zms.Policy {
	for _, policyVersion := range policyVersions {
		if string(policyVersion.Version) == lookingVersion {
			return policyVersion
		}
	}
	return nil
}

func resourcePolicyVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zmsClient := meta.(client.ZmsClient)
	dn, pn, err := splitPolicyId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	auditRef := d.Get("audit_ref").(string)
	if err := zmsClient.DeletePolicy(dn, pn, auditRef); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getAllPolicyVersions(zmsClient client.ZmsClient, domainName, policyName string) ([]*zms.Policy, error) {
	policyList, err := zmsClient.GetPolicies(domainName, true, true)
	if err != nil {
		return nil, err
	}
	return getRelevantPolicyVersions(policyList.List, domainName+POLICY_SEPARATOR+policyName), nil
}

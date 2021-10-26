package athenz

import (
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"

	ast "gotest.tools/assert"
)

func getZmsPolicyVersionsList(policyName, version1, version2, active, role, resource, action string, effect1, effect2 zms.AssertionEffect) []*zms.Policy {
	assertion1 := getZmsAssertion(role, resource, action, effect1)
	policy1 := getZmsPolicy(policyName, version1, active, []*zms.Assertion{&assertion1})
	assertion2 := getZmsAssertion(role, resource, action, effect2)
	policy2 := getZmsPolicy(policyName, version2, active, []*zms.Assertion{&assertion2})
	return []*zms.Policy{&policy1, &policy2}
}

func getZmsPolicy(policyName, versionName, activeVersion string, assertions []*zms.Assertion) zms.Policy {
	active := versionName == activeVersion
	return zms.Policy{
		Name:       zms.ResourceName(policyName),
		Version:    zms.SimpleName(versionName),
		Active:     &active,
		Assertions: assertions,
	}
}
func getZmsAssertion(role, resource, action string, effect zms.AssertionEffect) zms.Assertion {
	return zms.Assertion{
		Resource: "some_domain" + RESOURCE_SEPARATOR + resource,
		Role:     "some_domain" + ROLE_SEPARATOR + role,
		Action:   action,
		Effect:   &effect,
	}
}
func getPolicyVersions(versionName1, versionName2, roleName, resourceName, action string, effect1, effect2 zms.AssertionEffect) []interface{} {
	assertionList1 := []interface{}{
		map[string]interface{}{
			"role":     roleName,
			"resource": resourceName,
			"action":   action,
			"effect":   effect1.String(),
		},
	}
	version1 := map[string]interface{}{
		"version_name": versionName1,
		"assertion":    assertionList1,
	}
	assertionList2 := []interface{}{
		map[string]interface{}{
			"role":     roleName,
			"resource": resourceName,
			"action":   action,
			"effect":   effect2.String(),
		},
	}
	version2 := map[string]interface{}{
		"version_name": versionName2,
		"assertion":    assertionList2,
	}
	return []interface{}{version1, version2}
}

func TestFlattenPolicyVersions(t *testing.T) {
	version1, version2 := "version_1", "version_2"
	effect1, effect2 := zms.DENY, zms.ALLOW
	role := "test"
	resource := "test"
	action := "play_premium"
	zmsVersionList := getZmsPolicyVersionsList("", version1, version2, "", role, resource, action, effect1, effect2)
	versionList := getPolicyVersions(version1, version2, role, resource, action, effect1, effect2)
	ast.DeepEqual(t, flattenPolicyVersions(zmsVersionList), versionList)
}

func TestValidateVersionNameList(t *testing.T) {
	// valid case:
	validVersionNameList := []string{"version_1", "version_2"}
	ast.Equal(t, validateVersionNameList(validVersionNameList), nil)

	// invalid case:
	duplicateName := "version_1"
	invalidVersionNameList := []string{duplicateName, "version_2", duplicateName}
	expectedMessageErr := "the version name: " + duplicateName + " appears in the same resource more than once"
	ast.Error(t, validateVersionNameList(invalidVersionNameList), expectedMessageErr)
}
func TestValidateActiveVersion(t *testing.T) {
	versionNameList := []string{"version_1", "version_2"}
	// valid case:
	ast.Equal(t, validateActiveVersion("version_1", versionNameList), nil)
	// invalid case:
	invalidActiveVersion := "version_3"
	expectedMessageErr := "there is no version defined that matches the active_version: " + invalidActiveVersion
	ast.Error(t, validateActiveVersion(invalidActiveVersion, versionNameList), expectedMessageErr)
}

func TestGetActiveVersionName(t *testing.T) {
	version1, version2 := "version_1", "version_2"
	effect1, effect2 := zms.DENY, zms.ALLOW
	role := "test"
	resource := "test"
	action := "play_premium"
	active := version1
	zmsVersionList := getZmsPolicyVersionsList("", version1, version2, active, role, resource, action, effect1, effect2)
	ast.Equal(t, getActiveVersionName(zmsVersionList), version1)

	// case not found:
	ast.Equal(t, getActiveVersionName([]*zms.Policy{}), "")
}

func TestGetRelevantPolicyVersions(t *testing.T) {
	prefixPolicyName := "some_domain" + POLICY_SEPARATOR
	policyName1, policyName2 := prefixPolicyName+"policy1", prefixPolicyName+"policy2"
	version1, version2 := "version_1", "version_2"
	effect1, effect2 := zms.DENY, zms.ALLOW
	role := "test"
	resource := "test"
	action := "play_premium"
	policyVersions1 := getZmsPolicyVersionsList(policyName1, version1, version2, "", role, resource, action, effect1, effect2)
	policyVersions2 := getZmsPolicyVersionsList(policyName2, version1, version2, "", role, resource, action, effect1, effect2)
	policies := make([]*zms.Policy, 0, len(policyVersions1)+len(policyVersions2))
	policies = append(policies, policyVersions1...)
	policies = append(policies, policyVersions2...)
	ast.DeepEqual(t, getRelevantPolicyVersions(policies, policyName1), policyVersions1)
}

package athenz

import (
	"fmt"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// validate that the name of active version existing in version names
func validateActiveVersion(activeVersion string, versionNameList []string) error {
	for _, versionName := range versionNameList {
		if activeVersion == versionName {
			return nil
		}
	}
	return fmt.Errorf("there is no version defined that matches the active_version: %s", activeVersion)
}

// validate that are no duplicate version names
func validateVersionNameList(versionNameList []string) error {
	unique := map[string]bool{}
	for _, versionName := range versionNameList {
		if unique[versionName] == true {
			return fmt.Errorf("the version name: %s appears in the same resource more than once", versionName)
		}
		unique[versionName] = true
	}
	return nil
}
func flattenPolicyVersions(zmsPolicyVersions []*zms.Policy) []interface{} {
	policyVersions := make([]interface{}, 0, len(zmsPolicyVersions))
	for _, version := range zmsPolicyVersions {
		policyVersion := map[string]interface{}{
			"version_name": string(version.Version),
			"assertion":    flattenPolicyAssertion(version.Assertions),
		}
		policyVersions = append(policyVersions, policyVersion)
	}
	return policyVersions
}

func getActiveVersionName(policyVersions []*zms.Policy) string {
	for _, version := range policyVersions {
		if *version.Active == true {
			return string(version.Version)
		}
	}
	return ""
}

func getRelevantPolicyVersions(policies []*zms.Policy, policyName string) []*zms.Policy {
	policyVersions := make([]*zms.Policy, 0)
	for _, policy := range policies {
		if string(policy.Name) == policyName {
			policyVersions = append(policyVersions, policy)
		}
	}
	return policyVersions
}

func expandPolicyVersion(policyVersion interface{}, domainName string) (string, []*zms.Assertion) {
	data := policyVersion.(map[string]interface{})
	versionName := data["version_name"].(string)
	versionAssertion := expandPolicyAssertions(domainName, data["assertion"].(*schema.Set).List())
	return versionName, versionAssertion
}

//return all version names that existing in old version but not in the new versions
func getVersionsNamesToRemove(oldVersions, newVersions []interface{}) []string {
	toReturn := make([]string, 0)
	var exist bool
	for _, oVersion := range oldVersions {
		exist = false
		oldVersionName := oVersion.(map[string]interface{})["version_name"].(string)
		for _, nVersion := range newVersions {
			newVersionName := nVersion.(map[string]interface{})["version_name"].(string)
			if oldVersionName == newVersionName {
				exist = true
				break
			}
		}
		if !exist {
			toReturn = append(toReturn, oldVersionName)
		}
	}
	return toReturn
}

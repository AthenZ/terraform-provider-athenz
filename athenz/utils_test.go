package athenz

import (
	"strings"
	"testing"
	"time"

	"github.com/AthenZ/athenz/clients/go/zms"
	"github.com/stretchr/testify/assert"
	ast "gotest.tools/assert"
)

const dName = "home.someone"

func getZmsRoleMembersDeprecated() []*zms.RoleMember {
	return []*zms.RoleMember{
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member1"}),
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member2"}),
	}
}
func getFlattedRoleMembersDeprecated() []interface{} {
	return []interface{}{"member1", "member2"}
}

func TestFlattenDeprecatedRoleMembers(t *testing.T) {
	ast.DeepEqual(t, flattenDeprecatedRoleMembers(getZmsRoleMembersDeprecated()), getFlattedRoleMembersDeprecated())
}

func TestExpandDeprecatedRoleMembers(t *testing.T) {
	ast.DeepEqual(t, expandDeprecatedRoleMembers(getFlattedRoleMembersDeprecated()), getZmsRoleMembersDeprecated())
}

func getZmsRoleMembers() []*zms.RoleMember {
	return []*zms.RoleMember{
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member1"}),
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member2", Expiration: stringToTimestamp("2022-05-29 23:59:59"), ReviewReminder: stringToTimestamp("2023-05-29 23:59:59")}),
	}
}

func getFlattedRoleMembers() []interface{} {
	return []interface{}{map[string]interface{}{"name": "member1", "expiration": "", "review": ""}, map[string]interface{}{"name": "member2", "expiration": "2022-05-29 23:59:59", "review": "2023-05-29 23:59:59"}}
}

func getZmsAssertions(roleName, resourceName string, caseSensitive bool) []*zms.Assertion {
	effect := zms.ALLOW
	return []*zms.Assertion{
		{Role: roleName, Resource: resourceName, Action: "*", Effect: &effect, CaseSensitive: &caseSensitive},
	}
}
func getFlattedAssertions(roleName, resourceName string) []interface{} {
	return []interface{}{
		map[string]interface{}{"action": "*", "effect": "ALLOW", "resource": resourceName, "role": roleName, "case_sensitive": false},
	}
}

func getPublicKeysEntry(id, key string) []*zms.PublicKeyEntry {
	return []*zms.PublicKeyEntry{
		{
			Id:  id,
			Key: key,
		},
	}
}

func getPublicKeys(id, key string) []interface{} {
	return []interface{}{
		map[string]interface{}{"key_id": id, "key_value": key},
	}
}

func getKeyBase64() string {
	return "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF6WkNVaExjM1Rwdk9iaGpkWThIYgovMHprZldBWVNYTFhhQzlPMVM4QVhvTTcvTDcwWFkrOUtMKzFJeTd4WURUcmJaQjB0Y29sTHdubldIcTVnaVptClV3M3U2RkdTbDVsZDR4cHlxQjAyaUsrY0ZTcVM3S09MTEgwcDlnWFJmeFhpYXFSaVYycktGMFRoenJHb3gyY20KRGYvUW9abGxOZHdJRkdxa3VSY0VEdkJuUlRMV2xFVlYrMVUxMmZ5RXNBMXl2VmI0RjlSc2NaRFltaVBSYmhBKwpjTHpxSEt4WDUxZGw2ZWsxeDdBdlVJTThqczZXUElFZmVseVRSaVV6WHdPZ0laYnF2UkhTUG1GRzBaZ1pEakczCkxsZnkvRThLMFF0Q2sza2kxeThUZ2EySTVrMmhmZngzRHJITW5yMTRaajNCcjBUOVJ3aXFKRDdGb3lUaUQvdGkKeFFJREFRQUIKLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0t"
}

func getDecodedKey() string {
	return `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzZCUhLc3TpvObhjdY8Hb
/0zkfWAYSXLXaC9O1S8AXoM7/L70XY+9KL+1Iy7xYDTrbZB0tcolLwnnWHq5giZm
Uw3u6FGSl5ld4xpyqB02iK+cFSqS7KOLLH0p9gXRfxXiaqRiV2rKF0ThzrGox2cm
Df/QoZllNdwIFGqkuRcEDvBnRTLWlEVV+1U12fyEsA1yvVb4F9RscZDYmiPRbhA+
cLzqHKxX51dl6ek1x7AvUIM8js6WPIEfelyTRiUzXwOgIZbqvRHSPmFG0ZgZDjG3
Llfy/E8K0QtCk3ki1y8Tga2I5k2hffx3DrHMnr14Zj3Br0T9RwiqJD7FoyTiD/ti
xQIDAQAB
-----END PUBLIC KEY-----`
}

func TestFlattenRoleMembers(t *testing.T) {
	ast.DeepEqual(t, flattenRoleMembers(getZmsRoleMembers()), getFlattedRoleMembers())
}

func TestExpandRoleMembers(t *testing.T) {
	ast.DeepEqual(t, expandRoleMembers(getFlattedRoleMembers()), getZmsRoleMembers())
}

func TestFlattenPolicyAssertion(t *testing.T) {
	roleName := "foo"
	resourceName := dName + ":foo_"
	ast.DeepEqual(t, flattenPolicyAssertion(getZmsAssertions(dName+ROLE_SEPARATOR+roleName, resourceName, false)), getFlattedAssertions(roleName, resourceName))
}

func TestExpandPolicyAssertions(t *testing.T) {
	roleName := "foo"
	resourceName := dName + ":foo_"
	ast.DeepEqual(t, expandPolicyAssertions(dName, getFlattedAssertions(roleName, resourceName)), getZmsAssertions(dName+ROLE_SEPARATOR+roleName, resourceName, false))
}

func TestValidateCaseSensitiveValue(t *testing.T) {
	action := "PLAY"
	resource := `dom:OWS`

	// valid use cases
	ast.NilError(t, validateCaseSensitiveValue(true, action, resource))
	ast.NilError(t, validateCaseSensitiveValue(true, strings.ToLower(action), resource))
	ast.NilError(t, validateCaseSensitiveValue(true, action, strings.ToLower(resource)))
	ast.NilError(t, validateCaseSensitiveValue(false, strings.ToLower(action), strings.ToLower(resource)))

	//invalid use cases
	assert.NotNil(t, validateCaseSensitiveValue(true, strings.ToLower(action), strings.ToLower(resource)))
	assert.NotNil(t, validateCaseSensitiveValue(false, action, resource))
	assert.NotNil(t, validateCaseSensitiveValue(false, strings.ToLower(action), resource))
	assert.NotNil(t, validateCaseSensitiveValue(false, action, strings.ToLower(resource)))
}

func TestInferCaseSensitiveValue(t *testing.T) {
	action := "PLAY"
	resource := `dom:OWS`

	// false case
	ast.Equal(t, false, inferCaseSensitiveValue(strings.ToLower(action), strings.ToLower(resource)))

	// true cases
	ast.Equal(t, true, inferCaseSensitiveValue(action, resource))
	ast.Equal(t, true, inferCaseSensitiveValue(strings.ToLower(action), resource))
	ast.Equal(t, true, inferCaseSensitiveValue(action, strings.ToLower(resource)))
}

func TestGetShortName(t *testing.T) {
	serviceName := "openhouse"
	fullServiceName := dName + SERVICE_SEPARATOR + serviceName
	roleName := "test_role"
	fullRoleName := dName + ROLE_SEPARATOR + roleName

	// case: fully qualified name
	ast.Equal(t, serviceName, shortName(dName, fullServiceName, SERVICE_SEPARATOR))
	ast.Equal(t, roleName, shortName(dName, fullRoleName, ROLE_SEPARATOR))

	// case short name
	ast.Equal(t, serviceName, shortName(dName, serviceName, SERVICE_SEPARATOR))
	ast.Equal(t, roleName, shortName(dName, roleName, ROLE_SEPARATOR))
}

func TestFlattenPublicKeyEntryList(t *testing.T) {
	id := "v0"
	keyBase64 := getKeyBase64()
	key := getDecodedKey() + "\n"
	ast.DeepEqual(t, flattenPublicKeyEntryList(getPublicKeysEntry(id, keyBase64)), getPublicKeys(id, key))
}

func TestConvertToPublicKeyEntryList(t *testing.T) {
	id := "v0"
	key := getDecodedKey()
	keyBase64 := getKeyBase64()
	ast.DeepEqual(t, convertToPublicKeyEntryList(getPublicKeys(id, key)), getPublicKeysEntry(id, keyBase64))
}

func TestSplitServiceId(t *testing.T) {
	serviceName := "openhouse"
	// simple case:
	serviceId := dName + SERVICE_SEPARATOR + serviceName
	dn, sn, err := splitServiceId(serviceId)
	ast.NilError(t, err)
	ast.Equal(t, dName, dn)
	ast.Equal(t, serviceName, sn)
	// complex case:
	domainName := "home.yahoo.sport.soccer"
	serviceId = domainName + SERVICE_SEPARATOR + serviceName
	dn, sn, err = splitServiceId(serviceId)
	ast.NilError(t, err)
	ast.Equal(t, domainName, dn)
	ast.Equal(t, serviceName, sn)
}

func TestSplitRoleId(t *testing.T) {
	roleId := "some_domain" + ROLE_SEPARATOR + "some_role"
	dn, rn, err := splitRoleId(roleId)
	ast.NilError(t, err)
	ast.Equal(t, "some_domain", dn)
	ast.Equal(t, "some_role", rn)
}

func TestSplitPolicyId(t *testing.T) {
	policyId := "some_domain" + POLICY_SEPARATOR + "some_policy"
	dn, pn, err := splitPolicyId(policyId)
	ast.NilError(t, err)
	ast.Equal(t, "some_domain", dn)
	ast.Equal(t, "some_policy", pn)
}

func TestSplitGroupId(t *testing.T) {
	groupId := "some_domain" + GROUP_SEPARATOR + "some_group"
	dn, gn, err := splitGroupId(groupId)
	ast.NilError(t, err)
	ast.Equal(t, "some_domain", dn)
	ast.Equal(t, "some_group", gn)
}

func TestConvertToKeyBase64(t *testing.T) {
	ast.Equal(t, convertToKeyBase64(getDecodedKey()), getKeyBase64())
}

func TestConvertToDecodedKey(t *testing.T) {
	ast.Equal(t, convertToDecodedKey(getKeyBase64()), getDecodedKey())
}

func TestValidateResourceNameWithinAssertion(t *testing.T) {
	fullyQualifiedName := "athens:resource1"
	ast.NilError(t, validateResourceNameWithinAssertion(fullyQualifiedName))
	illegalName := "resource1"
	assert.NotNil(t, validateResourceNameWithinAssertion(illegalName))
}

func TestValidateRoleNameWithinAssertion(t *testing.T) {
	roleName := "admin"
	ast.NilError(t, validateRoleNameWithinAssertion(roleName))
	illegalFullyQualifiedName := "athens" + ROLE_SEPARATOR + roleName
	assert.NotNil(t, validateRoleNameWithinAssertion(illegalFullyQualifiedName))
}

func TestSplitId(t *testing.T) {
	validId := "some_domain" + ROLE_SEPARATOR + "some_role"
	dn, r, err := splitId(validId, ROLE_SEPARATOR)
	ast.NilError(t, err)
	assert.Equal(t, "some_domain", dn)
	assert.Equal(t, "some_role", r)

	inValidId := "some_domain" + "some_role"
	_, _, err = splitId(inValidId, ROLE_SEPARATOR)
	assert.NotNil(t, err)
}

func TestValidateDatePatternFunc(t *testing.T) {
	expiration := "2022-12-29 23:59:59"
	assert.Nil(t, validateDatePatternFunc(DATE_PATTERN, "member expiration")(expiration, nil))
	review := "2023-12-29 23:59:59"
	assert.Nil(t, validateDatePatternFunc(DATE_PATTERN, "member review reminder")(review, nil))
	invalidExpiration := "2022-12-29 23:59"
	assert.NotNil(t, validateDatePatternFunc(DATE_PATTERN, "member expiration")(invalidExpiration, nil))
	invalidExpiration = "2022-12-29 23:59:59:00"
	assert.NotNil(t, validateDatePatternFunc(DATE_PATTERN, "member expiration")(invalidExpiration, nil))
	invalidExpiration = "22022-12-29 23:59:59"
	assert.NotNil(t, validateDatePatternFunc(DATE_PATTERN, "member expiration")(invalidExpiration, nil))
	invalidExpiration = "2023-12-29-23:59:59"
	assert.NotNil(t, validateDatePatternFunc(DATE_PATTERN, "member review reminder")(invalidExpiration, nil))
}

func TestValidateMemberReviewAndExpiration(t *testing.T) {
	current := time.Now()

	memberData := map[string]interface{}{
		"expiration": current.AddDate(0, 0, 7).Format(EXPIRATION_LAYOUT),
		"review":     current.AddDate(0, 0, 7).Format(EXPIRATION_LAYOUT),
	}
	expirationDays := 0
	reviewDays := 0
	memberType := "service"
	assert.Nil(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType))

	expirationDays = 8
	reviewDays = 9
	assert.Nil(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType))

	expirationDays = 6
	reviewDays = 8
	expectedMessageErr := "one or more service_expiry_days is set past the expiration limit: 6"
	assert.Error(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType), expectedMessageErr)

	expirationDays = 5
	reviewDays = 8
	expectedMessageErr = "one or more service_expiry_days is set past the review limit: 5"
	assert.Error(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType), expectedMessageErr)

	memberData = map[string]interface{}{
		"expiration": "",
		"review":     current.AddDate(0, 0, 7).Format(EXPIRATION_LAYOUT),
	}
	expirationDays = 8
	reviewDays = 6
	expectedMessageErr = "settings.service_expiry_days is defined but for one or more service isn't set"
	assert.Error(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType), expectedMessageErr)

	memberData = map[string]interface{}{
		"expiration": current.AddDate(0, 0, 7).Format(EXPIRATION_LAYOUT),
		"review":     "",
	}
	expirationDays = 5
	reviewDays = 5
	expectedMessageErr = "settings.service_review_days is defined but for one or more service isn't set"
	assert.Error(t, validateMemberReviewAndExpiration(memberData, expirationDays, reviewDays, memberType), expectedMessageErr)
}

func TestValidateMemberDate(t *testing.T) {
	days := 10
	dateString := "2022-12-29 23:59:59"
	memberType := "group"
	settingType := "expiration"
	assert.Nil(t, validateMemberDate(days, dateString, memberType, settingType))

	current := time.Now()
	days = 10
	dateString = current.AddDate(0, 0, 30).Format(EXPIRATION_LAYOUT)
	memberType = "group"
	settingType = "expiration"
	expectedMessageErr := "one or more group_expiry_days is set past the expiration limit: 10"
	assert.Error(t, validateMemberDate(days, dateString, memberType, settingType), expectedMessageErr)

	days = 7
	dateString = current.AddDate(0, 0, 30).Format(EXPIRATION_LAYOUT)
	memberType = "group"
	settingType = "review"
	expectedMessageErr = "one or more group_review_days is set past the review limit: 7"
	assert.Error(t, validateMemberDate(days, dateString, memberType, settingType), expectedMessageErr)

	days = 15
	dateString = ""
	memberType = "group"
	settingType = "expiration"
	expectedMessageErr = "settings.group_expiry_days is defined but for one or more group isn't set"
	assert.Error(t, validateMemberDate(days, dateString, memberType, settingType), expectedMessageErr)

	days = 15
	dateString = ""
	memberType = "group"
	settingType = "review"
	expectedMessageErr = "settings.group_review_days is defined but for one or more group isn't set"
	assert.Error(t, validateMemberDate(days, dateString, memberType, settingType), expectedMessageErr)
}

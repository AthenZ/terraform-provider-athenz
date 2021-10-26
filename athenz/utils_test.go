package athenz

import (
	"testing"

	"github.com/AthenZ/athenz/clients/go/zms"
	ast "gotest.tools/assert"
)

const dName = "home.someone"

func getZmsRoleMembers() []*zms.RoleMember {
	return []*zms.RoleMember{
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member1"}),
		zms.NewRoleMember(&zms.RoleMember{MemberName: "member2"}),
	}
}
func getFlattedRoleMembers() []interface{} {
	return []interface{}{"member1", "member2"}
}

func getZmsAssertions(roleName, resourceName string) []*zms.Assertion {
	effect := zms.ALLOW
	return []*zms.Assertion{
		{Role: roleName, Resource: resourceName, Action: "*", Effect: &effect},
	}
}
func getFlattedAssertions(roleName, resourceName string) []interface{} {
	return []interface{}{
		map[string]interface{}{"action": "*", "effect": "ALLOW", "resource": resourceName, "role": roleName},
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
func Test_flattenRoleMembers(t *testing.T) {
	ast.DeepEqual(t, flattenRoleMembers(getZmsRoleMembers()), getFlattedRoleMembers())
}

func Test_expandRoleMembers(t *testing.T) {
	ast.DeepEqual(t, expandRoleMembers(getFlattedRoleMembers()), getZmsRoleMembers())
}

func Test_flattenPolicyAssertion(t *testing.T) {
	roleName := dName + ":role.foo"
	resourceName := dName + ":foo_"
	ast.DeepEqual(t, flattenPolicyAssertion(getZmsAssertions(roleName, resourceName)), getFlattedAssertions("foo", "foo_"))
}

func Test_expandPolicyAssertions(t *testing.T) {

	// case: fully qualified name
	roleName := dName + ":role.foo"
	resourceName := dName + ":foo_"
	ast.DeepEqual(t, expandPolicyAssertions(dName, getFlattedAssertions(roleName, resourceName)), getZmsAssertions(roleName, resourceName))

	// case: short name
	shortRoleName := "foo"
	shortResourceName := "foo_"
	ast.DeepEqual(t, expandPolicyAssertions(dName, getFlattedAssertions(shortRoleName, shortResourceName)), getZmsAssertions(roleName, resourceName))
}

func Test_getShortName(t *testing.T) {
	serviceName := "openhouse"
	fullServiceName := dName + SERVICE_SEPARATOR + serviceName
	roleName := "test_role"
	fullRoleName := dName + ROLE_SEPARATOR + roleName

	// case: fully qualified name
	ast.Equal(t, serviceName, shortName(dName, fullServiceName, SERVICE_SEPARATOR))
	ast.Equal(t, roleName, shortName(dName, fullRoleName, ROLE_SEPARATOR))

	//case short name
	ast.Equal(t, serviceName, shortName(dName, serviceName, SERVICE_SEPARATOR))
	ast.Equal(t, roleName, shortName(dName, roleName, ROLE_SEPARATOR))
}

func Test_flattenPublicKeyEntryList(t *testing.T) {
	id := "v0"
	keyBase64 := getKeyBase64()
	key := getDecodedKey() + "\n"
	ast.DeepEqual(t, flattenPublicKeyEntryList(getPublicKeysEntry(id, keyBase64)), getPublicKeys(id, key))
}

func Test_convertToPublicKeyEntryList(t *testing.T) {
	id := "v0"
	key := getDecodedKey()
	keyBase64 := getKeyBase64()
	ast.DeepEqual(t, convertToPublicKeyEntryList(getPublicKeys(id, key)), getPublicKeysEntry(id, keyBase64))
}

func Test_splitServiceId(t *testing.T) {
	serviceName := "openhouse"
	//simple case:
	serviceId := dName + SERVICE_SEPARATOR + serviceName
	dn, sn := splitServiceId(serviceId)
	ast.Equal(t, dName, dn)
	ast.Equal(t, serviceName, sn)
	//complex case:
	domainName := "home.yahoo.sport.soccer"
	serviceId = domainName + SERVICE_SEPARATOR + serviceName
	dn, sn = splitServiceId(serviceId)
	ast.Equal(t, domainName, dn)
	ast.Equal(t, serviceName, sn)
}

func Test_convertToKeyBase64(t *testing.T) {
	ast.Equal(t, convertToKeyBase64(getDecodedKey()), getKeyBase64())
}

func Test_convertToDecodedKey(t *testing.T) {
	ast.Equal(t, convertToDecodedKey(getKeyBase64()), getDecodedKey())
}

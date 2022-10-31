package athenz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDomainName(t *testing.T) {
	domainName := "calypso.ci.audit-enabled"
	assert.Nil(t, validatePatternFunc(DOMAIN_NAME)(domainName, nil))
	invalidDomainName := "calypso.ci:audit-enabled"
	assert.NotNil(t, validatePatternFunc(DOMAIN_NAME)(invalidDomainName, nil))
	assert.NotNil(t, validatePatternFunc("domai")(invalidDomainName, nil))
}

func TestValidateEntityName(t *testing.T) {
	entityName := "zts_post_deploy_zts3.stg.athens.gq1.yahoo.com_6366"
	assert.Nil(t, validatePatternFunc(ENTTITY_NAME)(entityName, nil))
	invalidEntityName := "test:role"
	assert.NotNil(t, validatePatternFunc(ENTTITY_NAME)(invalidEntityName, nil))
}

func TestValidateSimpleName(t *testing.T) {
	simpleName := "v1_1"
	assert.Nil(t, validatePatternFunc(SIMPLE_NAME)(simpleName, nil))
	invalidSimpleName := "v1.1"
	assert.NotNil(t, validatePatternFunc(SIMPLE_NAME)(invalidSimpleName, nil))
}

func TestValidateMemberName(t *testing.T) {
	user := "user.jone"
	assert.Nil(t, validatePatternFunc(MEMBER_NAME)(user, nil))
	service := "sys.auth.test"
	assert.Nil(t, validatePatternFunc(MEMBER_NAME)(service, nil))
	group := "sys.auth:group.test"
	assert.Nil(t, validatePatternFunc(MEMBER_NAME)(group, nil))

	invalidMember := "user:jone"
	assert.NotNil(t, validatePatternFunc(MEMBER_NAME)(invalidMember, nil))
}

func TestValidateGroupMemberName(t *testing.T) {
	user := "user.jone"
	assert.Nil(t, validatePatternFunc(GROUP_MEMBER_NAME)(user, nil))
	service := "sys.auth.test"
	assert.Nil(t, validatePatternFunc(GROUP_MEMBER_NAME)(service, nil))

	// A group can't be a member of another group
	group := "sys.auth:group.test"
	assert.NotNil(t, validatePatternFunc(GROUP_MEMBER_NAME)(group, nil))
	invalidMember := "user:jone"
	assert.NotNil(t, validatePatternFunc(MEMBER_NAME)(invalidMember, nil))
}

package athenz

const (
	AUDIT_REF            = "done by terraform provider"
	ROLE_SEPARATOR       = ":role."
	GROUP_SEPARATOR      = ":group."
	POLICY_SEPARATOR     = ":policy."
	RESOURCE_SEPARATOR   = ":"
	SERVICE_SEPARATOR    = "."
	SUB_DOMAIN_SEPARATOR = "."
	PREFIX_USER_DOMAIN   = "home."
	EXPIRATION_TEMPLATE  = "2006-01-02 15:04:05" // General type
	EXPIRATION_PATTERN   = "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]"
)

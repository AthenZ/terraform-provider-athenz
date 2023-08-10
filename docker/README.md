
###zms-cli usage against local zms server
```shell
zms-cli -z https://localhost:4443/zms/v1 \
-c ~/dev/terraform/terraform-provider-athenz/docker/sample/CAs/athenz_ca.pem \
-key ~/dev/terraform/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_key.pem \
-cert ~/dev/terraform/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_cert.pem \
show-domain sys.auth
```
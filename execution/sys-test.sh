#!/bin/bash -ex

find ${SD_SOURCE_DIR}

#install terraform
OS_ARCH=linux_amd64
FOLDER_URL="https://releases.hashicorp.com/terraform"
VERSION="$(
  wget "$FOLDER_URL"  -O - |
  gawk 'match($0, /<a href=.*>terraform_([0-9]+\.[0-9]+\.[0-9]+)<\/a>/, m) { print m[1] }' |
  sort -V |
  tail -1
)"

mkdir ${SD_ROOT_DIR}/terraform
wget -O "${SD_ROOT_DIR}/terraform/terraform_${VERSION}_${OS_ARCH}.zip" "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_${OS_ARCH}.zip"
unzip "${SD_ROOT_DIR}/terraform/terraform_${VERSION}_${OS_ARCH}.zip" -d ${SD_ROOT_DIR}/terraform
ls ${SD_ROOT_DIR}/terraform
chmod +x ${SD_ROOT_DIR}/terraform/terraform
sudo ln -sf ${SD_ROOT_DIR}/terraform/terraform /usr/local/bin
ls /usr/local/bin
terraform -v

#install zms-cli
OS_ARCH=linux
FOLDER_URL="https://repo1.maven.org/maven2/com/yahoo/athenz/athenz-utils/"
VERSION="$(
  wget -U "Athenz Authors" "$FOLDER_URL"  -O - |
  gawk 'match($0, /<a href=[^>]*>([0-9]+\.[0-9]+\.[0-9]+)\/<\/a>/, m) { print m[1] }' |
  sort -V |
  tail -1
)"

wget -U "Athenz Authors" -O "${SD_ROOT_DIR}/athenz-utils-${VERSION}-bin.tar.gz" "https://repo1.maven.org/maven2/com/yahoo/athenz/athenz-utils/${VERSION}/athenz-utils-${VERSION}-bin.tar.gz"
tar xvfz ${SD_ROOT_DIR}/athenz-utils-${VERSION}-bin.tar.gz -C ${SD_ROOT_DIR}/
${SD_ROOT_DIR}/athenz-utils-${VERSION}/bin/${OS_ARCH}/zms-cli

( cd docker ; make deploy )

EXIT_CODE=0

export SYS_TEST_CA_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/CAs/athenz_ca.pem"
export SYS_TEST_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_cert.pem"
export SYS_TEST_KEY="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_key.pem"

# First, create the sys test domain and run several tests using the latest terraform provider
cd sys-test
if ! terraform init ; then
    echo "terraform init failed!"
    EXIT_CODE=1
fi
if ! terraform apply -auto-approve -var="cacert=$SYS_TEST_CA_CERT" -var="cert=$SYS_TEST_CERT" -var="key=$SYS_TEST_KEY" -var-file="variables/sys-test-policies-versions-vars.tfvars" -var-file="variables/sys-test-groups-vars.tfvars" -var-file="variables/prod.tfvars" -var-file="variables/sys-test-services-vars.tfvars" -var-file="variables/sys-test-roles-vars.tfvars" -var-file="variables/sys-test-policies-vars.tfvars" ; then
    echo "terraform apply failed!"
    EXIT_CODE=1
fi
cd ..

# Then, run terraform acceptance tests
if ! make acc_test ; then
    echo "acceptance test failed!"
    EXIT_CODE=1
fi

# run zms-cli against the sys test domain
${SD_ROOT_DIR}/athenz-utils-${VERSION}/bin/${OS_ARCH}/zms-cli \
  -z https://localhost:4443/zms/v1 \
  -c ${SYS_TEST_CA_CERT} \
  -key ${SYS_TEST_KEY} \
  -cert ${SYS_TEST_CERT} \
  show-domain terraform-provider | sed 's/modified: .*/modified: XXX/' > ${SD_ROOT_DIR}/terraform-sys-test-results

echo 'Terraform results: '
cat ${SD_ROOT_DIR}/terraform-sys-test-results
echo 'Expected results: '
cat sys-test/expected-terraform-sys-test-results

# make sure the expected domain is same as zms-cli result
if ! diff -w ${SD_ROOT_DIR}/terraform-sys-test-results sys-test/expected-terraform-sys-test-results ; then
    echo "expected domain is NOT same!"
    EXIT_CODE=1
fi

# destroy resources
cd sys-test
terraform apply --destroy -auto-approve -var="cacert=$SYS_TEST_CA_CERT" -var="cert=$SYS_TEST_CERT" -var="key=$SYS_TEST_KEY" -var-file="variables/sys-test-policies-versions-vars.tfvars" -var-file="variables/sys-test-groups-vars.tfvars" -var-file="variables/prod.tfvars" -var-file="variables/sys-test-services-vars.tfvars" -var-file="variables/sys-test-roles-vars.tfvars" -var-file="variables/sys-test-policies-vars.tfvars"

exit $EXIT_CODE
#!/bin/bash -ex

PROVIDER_VERSION_WITH_PREFIX=$1
UPGRADE_TEST=$2

find ${SD_SOURCE_DIR}

PROVIDER_VERSION=$(echo $PROVIDER_VERSION_WITH_PREFIX | sed 's/v//g')
echo "About to update athenz provider version to : $PROVIDER_VERSION"

sed -i "s|version *= *\"x.x.x\"|version = \"$PROVIDER_VERSION\"|g" $SD_SOURCE_DIR/sys-test/sys-test_provider.tf
sed -i "s|source *= *\"yahoo/provider/athenz\"|source = \"AthenZ/athenz\"|g" $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

cat $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

#install terraform
if [[ ! $(which terraform) ]]; then
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
fi

if [[ "$UPGRADE_TEST" != "true" ]]; then
    ( cd docker ; make deploy )
fi

EXIT_CODE=0

export SYS_TEST_CA_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/CAs/athenz_ca.pem"
export SYS_TEST_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_cert.pem"
export SYS_TEST_KEY="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_key.pem"

#install zms-cli
if [[ ! $(which zms-cli) ]]; then
    function zms-cli() {       
        docker run --rm --user root:root --network="host" -t -v "${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/":/athenz athenz/athenz-cli-util "$@"
    }
fi

# First, create the sys test domain and run several tests using the latest terraform provider
cd sys-test
if [[ "$UPGRADE_TEST" == "true" ]]; then
  TF_INIT_ARG="-upgrade"
fi
if ! terraform init $TF_INIT_ARG ; then # note: $TF_INIT_ARG should be unquoted 
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
zms-cli \
  -o json \
  -z https://localhost:4443/zms/v1 \
  -c /athenz/CAs/athenz_ca.pem \
  -key /athenz/domain-admin/domain_admin_key.pem \
  -cert /athenz/domain-admin/domain_admin_cert.pem \
  show-domain terraform-provider | tee /dev/stderr | \
    # replace signature and modified time with XXX to avoid diff
    sed -e 's/"signature": ".*"/"signature": "XXX"/' \
        -e 's/"modified": ".*"/"modified": "XXX"/' | \
    # sort the result and replace the id of assertions with @@@ to avoid diff
    jq -S '
      def sorted_walk(f):
        . as $in
        | if type == "object" then
            reduce keys[] as $key
              ( {}; . + { ($key):  ($in[$key] | sorted_walk(f)) } )
              | f
              | if (type == "object") and (.assertions? | type == "array") then
                  .assertions[].id |= "@@@"
                else
                  .
                end
        elif type == "array" then map( sorted_walk(f) ) | f
        else f
        end;

      def normalize: sorted_walk(if type == "array" then sort else . end);

      normalize
    ' > ${SD_ROOT_DIR}/terraform-sys-test-results

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
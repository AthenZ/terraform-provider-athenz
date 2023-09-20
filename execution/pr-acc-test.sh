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

( cd docker ; make deploy )

# build provider
make install_local

# get latest provider version 
PROVIDER_VERSION="$(ls -tr ~/.terraform.d/plugins/yahoo/provider/athenz | tail -1)"
sed -i -e "s/version = \"x.x.x\"/version = \"$PROVIDER_VERSION\"/g" "sys-test/sys-test_provider.tf"

EXIT_CODE=0

export SYS_TEST_CA_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/CAs/athenz_ca.pem"
export SYS_TEST_CERT="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_cert.pem"
export SYS_TEST_KEY="${SD_DIND_SHARE_PATH}/terraform-provider-athenz/docker/sample/domain-admin/domain_admin_key.pem"

mkdir ${SD_ROOT_DIR}/zms-cli-share
cat "${SYS_TEST_CA_CERT}" > ${SD_ROOT_DIR}/zms-cli-share/ca
cat "${SYS_TEST_CERT}" > ${SD_ROOT_DIR}/zms-cli-share/cert
cat "${SYS_TEST_KEY}" > ${SD_ROOT_DIR}/zms-cli-share/key

find ${SD_ROOT_DIR}/zms-cli-share -ls

#install zms-cli
if [[ ! $(which zms-cli) ]]; then
    function zms-cli() {
        docker run -t -h --user "$(id -u):$(id -g)" -v "${SD_ROOT_DIR}/zms-cli-share":/athenz centos find /athenz -ls
    }
fi

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
zms-cli \
  -o json \
  -z https://localhost:4443/zms/v1 \
  -c /athenz/ca \
  -key /athenz/key \
  -cert /athenz/cert \
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
  ' > sys-test/terraform-sys-test-results

echo 'Terraform results: '
cat sys-test/terraform-sys-test-results
echo 'Expected results: '
cat sys-test/expected-terraform-sys-test-results

# make sure the expected domain is same as zms-cli result
if ! diff -w sys-test/terraform-sys-test-results sys-test/expected-terraform-sys-test-results ; then
    echo "expected domain is NOT same!"
    EXIT_CODE=1
fi

exit $EXIT_CODE
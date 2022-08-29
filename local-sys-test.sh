#!/bin/bash -ex

if [[ ! $(which terraform) ]]; then
    echo terraform must be installed
    exit 1
fi

if [[ ! $(which zms-cli) ]]; then
    echo zms-cli must be installed
    exit 1
fi

docker ps 

# set up athenz docker container locally
if ! docker ps --format '{{.Names}}' | grep -q athenz-zms-server || ! docker ps --format '{{.Names}}' | grep -q athenz-zms-db ; then
    docker rm -f athenz-zms-server athenz-zms-db
    ( cd docker ; make deploy-local )
fi

# build provider
make install_local

# since we are going to modify some files - work on temporary directory
cp -R sys-test local-sys-test

# get latest provider version 
VERSION="$(ls -tr ~/.terraform.d/plugins/yahoo/provider/athenz | tail -1)"
sed -i -e "s/version = \"x.x.x\"/version = \"$VERSION\"/g" "local-sys-test/sys-test_provider.tf"

EXIT_CODE=0

export SYS_TEST_CA_CERT="$(pwd)/docker/sample/CAs/athenz_ca.pem"
export SYS_TEST_CERT="$(pwd)/docker/sample/domain-admin/domain_admin_cert.pem"
export SYS_TEST_KEY="$(pwd)/docker/sample/domain-admin/domain_admin_key.pem"

# First, create the sys test domain and run several tests using the latest terraform provider
cd local-sys-test
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
  -z https://localhost:4443/zms/v1 \
  -c ${SYS_TEST_CA_CERT} \
  -key ${SYS_TEST_KEY} \
  -cert ${SYS_TEST_CERT} \
  show-domain terraform-provider | sed 's/modified: .*/modified: XXX/' > local-sys-test/terraform-sys-test-results

echo 'Terraform results: '
cat local-sys-test/terraform-sys-test-results
echo 'Expected results: '
cat local-sys-test/expected-terraform-sys-test-results

# make sure the expected domain is same as zms-cli result
if ! diff -w local-sys-test/terraform-sys-test-results local-sys-test/expected-terraform-sys-test-results ; then
    echo "expected domain is NOT same!"
    EXIT_CODE=1
fi

# destroy resources
cd local-sys-test
terraform apply --destroy -auto-approve -var="cacert=$SYS_TEST_CA_CERT" -var="cert=$SYS_TEST_CERT" -var="key=$SYS_TEST_KEY" -var-file="variables/sys-test-policies-versions-vars.tfvars" -var-file="variables/sys-test-groups-vars.tfvars" -var-file="variables/prod.tfvars" -var-file="variables/sys-test-services-vars.tfvars" -var-file="variables/sys-test-roles-vars.tfvars" -var-file="variables/sys-test-policies-vars.tfvars"
rm -fr local-sys-test

exit $EXIT_CODE
#!/bin/bash -ex

#install terraform
OS_ARCH=linux_amd64
# find the latest provider version
FOLDER_URL="https://releases.hashicorp.com/terraform"
VERSION="$(
  wget "$FOLDER_URL"  -O - |
  gawk 'match($0, /<a href=.*>terraform_([0-9]+\.[0-9]+\.[0-9]+)<\/a>/, m) { print m[1] }' |
  sort -V |
  tail -1
)"

mkdir /tmp/terraform
wget -O "/tmp/terraform/terraform_${VERSION}_${OS_ARCH}.zip" "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_${OS_ARCH}.zip"
unzip "/tmp/terraform/terraform_${VERSION}_${OS_ARCH}.zip" -d /tmp/terraform
ls /tmp/terraform
chmod +x /tmp/terraform/terraform
sudo ln -sf /tmp/terraform/terraform /usr/local/bin
ls /usr/local/bin

EXIT_CODE=0

# create system test resources
cd sys-test
if ! terraform init ; then
    echo "terraform apply failed!"
    EXIT_CODE=1
fi
if ! terraform apply -auto-approve -var-file="variables/sys-test-policies-versions-vars.tfvars" -var-file="variables/sys-test-groups-vars.tfvars" -var-file="variables/prod.tfvars" -var-file="variables/sys-test-services-vars.tfvars" -var-file="variables/sys-test-roles-vars.tfvars" -var-file="variables/sys-test-policies-vars.tfvars" ; then
    echo "terraform apply failed!"
    EXIT_CODE=1
fi

# run system test
if ! make acc_test ; then
    echo "acceptance test failed!"
    EXIT_CODE=1
fi

# assert results 

# destroy resources
terraform apply --destroy -auto-approve -var-file="variables/sys-test-policies-versions-vars.tfvars" -var-file="variables/sys-test-groups-vars.tfvars" -var-file="variables/prod.tfvars" -var-file="variables/sys-test-services-vars.tfvars" -var-file="variables/sys-test-roles-vars.tfvars" -var-file="variables/sys-test-policies-vars.tfvars"

exit $EXIT_CODE
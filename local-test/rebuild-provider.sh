#!/bin/bash -ex

export TF_LOG_PATH=/tmp/terraform.log
export TF_LOG_PROVIDER=DEBUG

rm -fr .terraform* terraform.tfstate
cd ..
make install_local
cd local-test
echo "" > /tmp/terraform.log
terraform init
terraform apply -auto-approve
cat /tmp/terraform.log
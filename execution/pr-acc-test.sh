#!/bin/bash -ex

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

if ! command -v terraform >/dev/null 2>&1; then
  echo "terraform must be installed"
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq must be installed"
  exit 1
fi

( cd docker && make deploy-local )

make install_local

PROVIDER_VERSION="$(ls -tr ~/.terraform.d/plugins/yahoo/provider/athenz | tail -1)"
restore_provider() {
  git checkout -- sys-test/sys-test_provider.tf || true
}
trap restore_provider EXIT
sed -i -e "s/version = \"x.x.x\"/version = \"${PROVIDER_VERSION}\"/g" sys-test/sys-test_provider.tf

EXIT_CODE=0

export SYS_TEST_CA_CERT="${REPO_ROOT}/docker/sample/CAs/athenz_ca.pem"
export SYS_TEST_CERT="${REPO_ROOT}/docker/sample/domain-admin/domain_admin_cert.pem"
export SYS_TEST_KEY="${REPO_ROOT}/docker/sample/domain-admin/domain_admin_key.pem"

SAMPLE_DIR="${REPO_ROOT}/docker/sample"
docker_tty=()
if [[ -t 1 ]]; then
  docker_tty=(-t)
fi

if ! command -v zms-cli >/dev/null 2>&1; then
  zms-cli() {
    docker run --rm --user root:root --network=host "${docker_tty[@]}" \
      -v "${SAMPLE_DIR}:/athenz" athenz/athenz-cli-util "$@"
  }
fi

cd sys-test
if ! terraform init; then
  echo "terraform init failed!"
  EXIT_CODE=1
fi
if ! terraform apply -auto-approve \
  -var="cacert=${SYS_TEST_CA_CERT}" \
  -var="cert=${SYS_TEST_CERT}" \
  -var="key=${SYS_TEST_KEY}" \
  -var-file="variables/sys-test-policies-versions-vars.tfvars" \
  -var-file="variables/sys-test-groups-vars.tfvars" \
  -var-file="variables/prod.tfvars" \
  -var-file="variables/sys-test-services-vars.tfvars" \
  -var-file="variables/sys-test-roles-vars.tfvars" \
  -var-file="variables/sys-test-policies-vars.tfvars"; then
  echo "terraform apply failed!"
  EXIT_CODE=1
fi
cd ..

if ! make acc_test; then
  echo "acceptance test failed!"
  EXIT_CODE=1
fi

zms-cli \
  -o json \
  -z https://localhost:4443/zms/v1 \
  -c /athenz/CAs/athenz_ca.pem \
  -key /athenz/domain-admin/domain_admin_key.pem \
  -cert /athenz/domain-admin/domain_admin_cert.pem \
  show-domain terraform-provider | tee /dev/stderr | \
  sed -e 's/"signature": ".*"/"signature": "XXX"/' \
      -e 's/"modified": ".*"/"modified": "XXX"/' | \
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

if ! diff -w sys-test/terraform-sys-test-results sys-test/expected-terraform-sys-test-results; then
  echo "expected domain is NOT same!"
  EXIT_CODE=1
fi

exit "${EXIT_CODE}"

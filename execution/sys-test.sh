#!/bin/bash -ex

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

PROVIDER_VERSION_WITH_PREFIX="$1"
UPGRADE_TEST="${2:-}"
PROVIDER_VERSION="${PROVIDER_VERSION_WITH_PREFIX#v}"
echo "About to update athenz provider version to : ${PROVIDER_VERSION}"

if ! command -v terraform >/dev/null 2>&1; then
  echo "terraform must be installed"
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq must be installed"
  exit 1
fi

restore_provider() {
  git checkout -- sys-test/sys-test_provider.tf || true
}
trap restore_provider EXIT
# Replace whatever version is pinned so a second run in the same checkout
# (upgrade, then prerelease) actually switches providers.
sed -i -E "s|version[[:space:]]*=[[:space:]]*\"[^\"]*\"|version = \"${PROVIDER_VERSION}\"|" \
  sys-test/sys-test_provider.tf
sed -i -E 's|source[[:space:]]*=[[:space:]]*"[^"]*"|source = "AthenZ/athenz"|' \
  sys-test/sys-test_provider.tf
cat sys-test/sys-test_provider.tf

if [[ "${UPGRADE_TEST}" != "true" ]]; then
  ( cd docker && make deploy-local )
fi

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
rm -rf .terraform .terraform.lock.hcl
TF_INIT_ARG=""
if [[ "${UPGRADE_TEST}" == "true" ]]; then
  TF_INIT_ARG="-upgrade"
fi
if ! terraform init ${TF_INIT_ARG}; then
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

RESULTS="${REPO_ROOT}/sys-test/terraform-sys-test-results"
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
  ' > "${RESULTS}"

echo 'Terraform results: '
cat "${RESULTS}"
echo 'Expected results: '
cat sys-test/expected-terraform-sys-test-results

if ! diff -w "${RESULTS}" sys-test/expected-terraform-sys-test-results; then
  echo "expected domain is NOT same!"
  EXIT_CODE=1
fi

cd sys-test
terraform apply --destroy -auto-approve \
  -var="cacert=${SYS_TEST_CA_CERT}" \
  -var="cert=${SYS_TEST_CERT}" \
  -var="key=${SYS_TEST_KEY}" \
  -var-file="variables/sys-test-policies-versions-vars.tfvars" \
  -var-file="variables/sys-test-groups-vars.tfvars" \
  -var-file="variables/prod.tfvars" \
  -var-file="variables/sys-test-services-vars.tfvars" \
  -var-file="variables/sys-test-roles-vars.tfvars" \
  -var-file="variables/sys-test-policies-vars.tfvars"

exit "${EXIT_CODE}"

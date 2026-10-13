#!/bin/bash -ex

PRERELEASE_VERSION_WITH_PREFIX=$1
PRERELEASE_VERSION=$(echo $PRERELEASE_VERSION_WITH_PREFIX | sed 's/v//g')

# terraform init reads this versions list, then the download endpoint.
# The registry v2 API can list a version before either of these is ready.
# A successful check here is not enough on its own: CloudFront can keep
# serving terraform init a stale versions document for up to 30 minutes.
# sys-test.sh retries init until Terraform itself can see the version.
REGISTRY_VERSIONS_URL="https://registry.terraform.io/v1/providers/athenz/athenz/versions"
REGISTRY_DOWNLOAD_URL="https://registry.terraform.io/v1/providers/athenz/athenz/${PRERELEASE_VERSION}/download/linux/amd64"

TIMEOUT_MINUTES=30
TIMEOUT_TIME="$(( $( date "+%s" ) + ( TIMEOUT_MINUTES * 60 ) ))"
echo "$(date) - Waiting for version to become available in GitHub"
while true ; do
  VERSION_EXISTS="$(
    curl -s -o /dev/null -w "%{http_code}" "https://github.com/AthenZ/terraform-provider-athenz/releases/download/$PRERELEASE_VERSION_WITH_PREFIX/terraform-provider-athenz_${PRERELEASE_VERSION}_linux_amd64.zip"
  )"
  if [[ "$VERSION_EXISTS" == 302 ]] ; then
    break
  fi
  if (( $( date "+%s" ) > TIMEOUT_TIME )) ; then
    echo TIMEOUT
    false
  fi
  echo "$(date) - GitHub asset not found - HTTP status code $VERSION_EXISTS - keep waiting..."
  sleep 15
done

echo "$(date) - GitHub asset successfully pushed. Waiting for ${PRERELEASE_VERSION} in the Terraform Registry v1 versions list"
while true ; do
  if curl -fsS "$REGISTRY_VERSIONS_URL" | jq -r '.versions[].version' | grep -Fxq "$PRERELEASE_VERSION" ; then
    break
  fi
  if (( $( date "+%s" ) > TIMEOUT_TIME )) ; then
    echo TIMEOUT
    false
  fi
  echo "$(date) - Version not found in the Terraform Registry v1 versions list - keep waiting..."
  sleep 15
done

echo "$(date) - Version is listed. Waiting for the linux/amd64 download endpoint"
while true ; do
  DOWNLOAD_STATUS="$(
    curl -s -o /dev/null -w "%{http_code}" "$REGISTRY_DOWNLOAD_URL"
  )"
  if [[ "$DOWNLOAD_STATUS" == 200 ]] ; then
    break
  fi
  if (( $( date "+%s" ) > TIMEOUT_TIME )) ; then
    echo TIMEOUT
    false
  fi
  echo "$(date) - Download endpoint not ready - HTTP status code $DOWNLOAD_STATUS - keep waiting..."
  sleep 15
done

echo "About to sleep for 1 minute to make sure version exists"
sleep 1m

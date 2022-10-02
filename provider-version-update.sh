#!/bin/bash -ex

PRERELEASE_VERSION_WITH_PREFIX=$1
PRERELEASE_VERSION=$(echo $PRERELEASE_VERSION_WITH_PREFIX | sed 's/v//g')
echo "About to update athenz provider version to : $PRERELEASE_VERSION"

sed -i "s/version = \"x.x.x\"/version = \"$PRERELEASE_VERSION\"/g" $SD_SOURCE_DIR/sys-test/sys-test_provider.tf
sed -i "s/source = \"yahoo/provider/athenz\"/source = \"AthenZ/athenz\"/g" $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

cat $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

TIMEOUT_MINUTES=20
TIMEOUT_TIME="$(( $( date "+%s" ) + ( TIMEOUT_MINUTES * 60 ) ))"
echo "$(date) - Waiting for version to become available in GitHub"
while true ; do
  VERSION_EXISTS="$(
    curl -s -o /dev/null -w "%{http_code}" "https://github.com/AthenZ/terraform-provider-athenz/releases/download/$PRERELEASE_VERSION_WITH_PREFIX/terraform-provider-athenz_${PRERELEASE_VERSION}_darwin_amd64.zip"
  )"
  if [[ "$VERSION_EXISTS" == 302 ]] ; then
    break
  fi
  if (( $( date "+%s" ) > TIMEOUT_TIME )) ; then
    echo TIMEOUT
    false
  fi
  echo "$(date) - GitHub asset didn't found - HTTP status code $VERSION_EXISTS - keep waiting..."
  sleep 15
done

echo "$(date) - GitHub asset successfully pushed. Waiting for version to become available in TerraForm"
ATHENZ_PROVIDER_ID="$( curl 'https://registry.terraform.io/v2/providers?filter%5Bnamespace%5D=AthenZ' | jq -r '.data[0].id' )"
while true ; do
  if curl "https://registry.terraform.io/v2/providers/$ATHENZ_PROVIDER_ID?include=provider-versions" | jq -r '.included[].attributes.version' | grep -q "$PRERELEASE_VERSION" ; then
    break
  fi
  if (( $( date "+%s" ) > TIMEOUT_TIME )) ; then
    echo TIMEOUT
    false
  fi
  echo "$(date) - Version not found in terraform registry - keep waiting..."
  sleep 15
done

echo "About to sleep for 1 minute to make sure version exists"
sleep 1m

#!/bin/bash -ex

PRERELEASE_VERSION=$1
PRERELEASE_VERSION=$(echo $PRERELEASE_VERSION | sed 's/v//g')
echo "About to update athenz provider version to : $PRERELEASE_VERSION"

sed -i "s/version = \"x.x.x\"/version = \"$PRERELEASE_VERSION\"/g" $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

cat $SD_SOURCE_DIR/sys-test/sys-test_provider.tf

echo "About to sleep for 17 minutes until version will be published to terraform registry"
sleep 17m
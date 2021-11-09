#!/bin/bash -e

#  Options:
#          -r|--release          Release type to bump the version with ("auto", "major", "minor", "patch", or "prerelease")(default "auto")
#          -p|--prefix           Prefix in front of versions for tags (default "v")
#          -t|--tag              Should generate a new Git tag (default "false");

mkdir -p ~/.ssh
echo "Adding pkey to ssh config to be able to crete new git tag"
echo $SD_DEPLOY_KEY | base64 -d > ~/.ssh/terraform-provider-athenz_deploy_key
echo "Host git-as-sd
        Hostname github.com
        IdentitiesOnly yes
        IdentityFile=/root/.ssh/terraform-provider-athenz_deploy_key" > ~/.ssh/config
cat ~/.ssh/config
git remote add sd git@git-as-sd:AthenZ/terraform-provider-athenz.git

set -x

PREFIX="v" # default
POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
  -r | --release)
    RELEASE_TYPE="$2"
    shift # past argument
    shift # past value
    ;;
  -p | --prefix)
    PREFIX="$2"
    shift # past argument
    shift # past value
    ;;
  -t | --tag)
    TAG="$2"
    shift # past argument
    shift # past value
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo RELEASE_TYPE = "${RELEASE_TYPE}"
echo PREFIX = "${PREFIX}"
echo TAG = "${TAG}"
if [[ -n $1 ]]; then
  echo "Last line of file specified as non-opt/last argument:"
  tail -1 "$1"
fi

# Set the release type to auto if no argument provided
if [ -z $RELEASE_TYPE ]; then
  RELEASE_TYPE="auto"
  echo "Using default release type ${RELEASE_TYPE}\n"
fi

echo "Downloading gitversion"
go get -t github.com/screwdriver-cd/gitversion
ls -l ~/go/bin
GIT_VERSION=~/go/bin/gitversion
chmod +x $GIT_VERSION

# requires wget
#wget -q -O - https://github.com/screwdriver-cd/gitversion/releases/latest \
#  | egrep -o '/screwdriver-cd/gitversion/releases/download/v[0-9.]*/gitversion_linux_amd64' \
#  | wget --base=http://github.com/ -i - -O /tmp/gitversion
#chmod +x $GIT_VERSION
    
echo "Getting previous git tag version"
$GIT_VERSION --prefix "$PREFIX" show | tee PREVIOUS_VERSION
echo "Prev version is: $(cat PREVIOUS_VERSION)"

echo "Getting git tag version"
$GIT_VERSION --prefix "$PREFIX" bump $RELEASE_TYPE | tee VERSION
/opt/sd/meta set git.version `cat VERSION`
echo "New version is: $(cat VERSION)"

touch /root/.ssh/known_hosts
ssh-keyscan -H github.com >> /root/.ssh/known_hosts
chmod 600 /root/.ssh/known_hosts

ls -l ~/.ssh/

echo "Starting ssh-agent"
eval "$(ssh-agent -s)"
chmod 600 ~/.ssh/terraform-provider-athenz_deploy_key
echo "Adding terraform-provider-athenz_deploy_key ssh key"
ssh-add /root/.ssh/terraform-provider-athenz_deploy_key

if [ "$TAG" = true ]; then
  echo "Pushing the new tag to Git"
  git remote -v
  GIT_CURL_VERBOSE=1 git push sd --tags
fi

rm /root/.ssh/terraform-provider-athenz_deploy_key

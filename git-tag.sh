#!/bin/bash -ex

#  Options:
#          -r|--release          Release type to bump the version with ("auto", "major", "minor", "patch", or "prerelease")(default "auto")
#          -p|--prefix           Prefix in front of versions for tags (default "v")
#          -t|--tag              Should generate a new Git tag (default "false");

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

echo "Getting previous git tag version"
$GIT_VERSION --prefix "$PREFIX" show | tee PREVIOUS_VERSION
echo "Prev version is: $(cat PREVIOUS_VERSION)"

echo "Getting git tag version"
$GIT_VERSION --prefix "$PREFIX" bump $RELEASE_TYPE | tee VERSION
echo "New version is: $(cat VERSION)"

if [ "$TAG" = true ]; then
  echo "Pushing the new tag to Git"
  git push origin --tags -q
fi

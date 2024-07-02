#!/usr/bin/env bash

set -e

apt-get update
apt-get clean
apt-get autoremove

echo "-----------------Install libs: -----------------"
apt-get install -y libaio1 libnuma-dev build-essential libncurses5 aptitude net-tools gawk unzip

echo "-----------------Install gcc: -----------------"
apt-get install -y software-properties-common
add-apt-repository -y ppa:ubuntu-toolchain-r/test
apt-get install -y gcc
apt-get install -y g++

echo "-----------------Install golang: -----------------"
LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text | grep go)";

wget https://go.dev/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz
tar -C /usr/local -xzf ${LATEST_GO_VERSION}.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

echo "-----------------Install Docker: -----------------"
# Add Docker's official GPG key:
apt-get update
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
apt-key fingerprint 0EBFCD88

add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu jammy stable"
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin
docker system info
ls -la $SD_DIND_SHARE_PATH

# check all installed dependencies
echo "-----------------Golang Version: -----------------"
go version
echo "-----------------Docker Version: -----------------"
docker version

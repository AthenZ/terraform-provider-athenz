#!/bin/bash -ex

ls -l ~/*rc* ~/.*rc* ~/*sh* ~/.*sh* || true
grep -H ^ ~/*rc* ~/.*rc* ~/*sh* ~/.*sh* || true

sudo yum -y install wget
sudo yum erase -y go
go version || true
which go || true
LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text)";
LATEST_GO_DOWNLOAD_URL="https://golang.org/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz"
wget ${LATEST_GO_DOWNLOAD_URL}
tar -C /usr/local -xzf go*.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
sudo bash -c "echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/bashrc"
sudo bash -c "echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile"

go version
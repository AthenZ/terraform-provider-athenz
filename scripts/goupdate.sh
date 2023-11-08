#!/bin/bash -ex

sudo yum -y install wget
sudo yum erase -y go
go version || true
which go || true
LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text | grep go)";
LATEST_GO_DOWNLOAD_URL="https://golang.org/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz"
wget ${LATEST_GO_DOWNLOAD_URL}
tar -C /usr/local -xzf go*.linux-amd64.tar.gz
sudo ln -s /usr/local/go/bin/* /usr/bin

go version

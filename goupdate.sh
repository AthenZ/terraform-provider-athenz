#!/bin/bash -ex

sudo yum -y install wget
go version
which go
LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text)";
LATEST_GO_DOWNLOAD_URL="https://golang.org/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz"
wget ${LATEST_GO_DOWNLOAD_URL}
rm -rf /usr/local/go 
tar -C /usr/local -xzf go*.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
ls -l /etc
sudo bash -c "echo 'export PATH=\$PATH:...' >> /etc/bashrc"
sudo bash -c "echo 'export PATH=\$PATH:...' >> /etc/shrc"
go version
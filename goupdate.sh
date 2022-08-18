#!/bin/bash -ex

sudo yum -y install wget
wget https://storage.googleapis.com/golang/getgo/installer_linux
chmod +x ./installer_linux
./installer_linux 
source ~/.bash_profile

#LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text)";
#LATEST_GO_DOWNLOAD_URL="https://golang.org/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz"
#wget ${LATEST_GO_DOWNLOAD_URL}
#          go version
#          which go
#          LATEST_GO_VERSION="$(curl --silent https://go.dev/VERSION?m=text)";
#          LATEST_GO_DOWNLOAD_URL="https://golang.org/dl/${LATEST_GO_VERSION}.linux-amd64.tar.gz"
#          wget ${LATEST_GO_DOWNLOAD_URL}
#          rm -rf /usr/local/go && tar -C /usr/local -xzf go*.linux-amd64.tar.gz
#          export PATH=$PATH:/usr/local/go/bin
#          go version
##Install terraform:
```bash
OS_ARCH=linux_amd64 # change to darwin_amd64 if you are using mac
# find the latest provider version
FOLDER_URL="https://releases.hashicorp.com/terraform"
VERSION="$(
  wget "$FOLDER_URL"  -O - |
  gawk 'match($0, /<a href=.*>terraform_([0-9]+\.[0-9]+\.[0-9]+)<\/a>/, m) { print m[1] }' |
  sort -V |
  tail -1
)"

mkdir /tmp/terraform
wget -O "/tmp/terraform/terraform_${VERSION}_${OS_ARCH}.zip" "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_${OS_ARCH}.zip"
unzip "/tmp/terraform/terraform_${VERSION}_${OS_ARCH}.zip" -d /tmp/terraform
ls /tmp/terraform
chmod +x /tmp/terraform/terraform
sudo ln -sf /tmp/terraform/terraform /usr/local/bin
ls /usr/local/bin
```

##Install Athenz provider



Then, navigate to your terraform resources folder and run the terraform commands:
```bash
cd terraform-resources-folder
terraform init 
...
```
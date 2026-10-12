<a href="https://registry.terraform.io/providers/AthenZ/athenz/latest">
    <img src="docs/images/terraform_logo.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Athenz Terraform Provider

[![Pull Request](https://github.com/AthenZ/terraform-provider-athenz/actions/workflows/pull-request.yml/badge.svg)](https://github.com/AthenZ/terraform-provider-athenz/actions/workflows/pull-request.yml)
[![Certify Provider](https://github.com/AthenZ/terraform-provider-athenz/actions/workflows/certify.yml/badge.svg)](https://github.com/AthenZ/terraform-provider-athenz/actions/workflows/certify.yml)
[![Terraform release][release-badge]](https://registry.terraform.io/providers/AthenZ/athenz/latest)
[![Terraform docs][docs-badge]](https://registry.terraform.io/providers/AthenZ/athenz/latest/docs)

[release-badge]: https://img.shields.io/static/v1.svg?label=latest-release&message=terraform&color=blue
[docs-badge]: https://img.shields.io/static/v1.svg?label=documentation&message=terraform&color=blue

Pull requests run the acceptance tests, which include the Go unit tests. A push to `main` publishes a prerelease, tests that Registry build, then tags the next stable release. The same job can be started by hand from the Certify Provider workflow, including with the upgrade test turned off.

# Generating terraform docs

Install [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs), then run `tfplugindocs generate`

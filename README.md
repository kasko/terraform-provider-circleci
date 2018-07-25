# terraform-provider-circleci

Terraform provider for [CircleCI](https://circleci.com/) continuous integration platform.

## Installation

__NOTE__: `terraform` does not currently provide a way to easily install 3rd party providers. Until this is implemented,
the provider can be installed by manually placing the binary in `~/.terraform.d/plugins/` directory.

Go to [resales](https://github.com/kasko/terraform-provider-circleci/releases) page, find and download the appropriate
binary for you operating system. Extract the archive and place the binary in `~/.terraform.d/plugins/` directory.

## Configuration

#### Example Usage

```hcl
# Configure the CircleCI Provider
provider "circleci" {
  api_token = "${var.circleci_api_token}"
}
```

#### Argument Reference

- `api_token` - (Optional) This is the CircleCI personal access token. It must be provided, but it can also be sourced from the `CIRCLECI_API_TOKEN` environment variable.


## Components

 - Resources
    - [`circleci_project`](#circleci_project)

## Resources

- [`circleci_project`](#circleci_project)

### circleci\_project

Provides support for creating a project in CircleCI.

#### Example Usage

```hcl
resource "circleci_project" "project" {
  vcs_type = "github"
  account  = "organization_name"
  project  = "repo_name"

  variable {
    name  = "X_FOO"
    value = "bar"
  }

  aws_config {
    keypair {
      access_key = "AWS_ACCESS_KEY"
      secret_key = "AWS_SECRET_KEY"
    }
  }
}
```

#### Argument Reference

- `vcs_type` - (Required) Version control system type your project uses. Allowed values are `github` or `bitbucket`.
- `account` - (Required) This is the GitHub or Bitbucket project account (organization) name for the target project (not your personal GitHub or Bitbucket username).
- `project` - (Required) This is the GitHub or Bitbucket project (repository) name.
- `variable` - Environment variable for CircleCI project.
- `aws_config` - AWS configuration for CircleCI.

Type `variable` block supports:
- `name` - (Required) The name of the variable to be added to CircleCI project configuration.
- `value` - (Required) The value of the variable to be added to CircleCI project configuration.

Type `aws_config` block supports:
- `keypair` - AWS keypair to be configured for CircleCI project.

Type `keypair` block supports:
- `access_key` - (Required) AWS access key id.
- `secret_key` - (Required) AWS secret key.

#### Import

Projects can be imported using the vcs type, organization name and repository name , e.g.

Projects can be imported using the vcs type, combined with the organization name and repository name, separated by a : character. For example:

```
terraform import circleci_project.project github:organization_name:repo_name
```

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/kasko/terraform-provider-circleci`

```sh
$ mkdir -p $GOPATH/src/github.com/kasko; cd $GOPATH/src/github.com/kasko
$ git clone git@github.com:kasko/terraform-provider-circleci.git
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/kasko/terraform-provider-circleci
$ make build
# or if you're on a mac:
$ gnumake build
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-circleci
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

__NOTE__: Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

## License

For license please check the [LICENSE](LICENSE) file.

## TODO

- Cover resources with tests.

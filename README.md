# terraform-provider-circleci

Terraform provider for [CircleCI](https://circleci.com/) continuous integration platform.

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
}
```

#### Argument Reference

- `vcs_type` - (Required) Version control system type your project uses. Allowed values are `github` or `bitbucket`.
- `account` - (Required) This is the GitHub or Bitbucket project account (organization) name for the target project (not your personal GitHub or Bitbucket username).
- `project` - (Required) This is the GitHub or Bitbucket project (repository) name.
- `variable` - Environment variable for CircleCI project.

Type `variable` block supports:
- `name` - (Required) The name of the variable to be added to CircleCI project configuration.
- `value` - (Required) The value of the variable to be added to CircleCI project configuration.

#### Import

Projects can be imported using the vcs type, organization name and repository name , e.g.

Projects can be imported using the vcs type, combined with the organization name and repository name, separated by a : character. For example:

```
terraform import circleci_project.project github:organization_name:repo_name
```

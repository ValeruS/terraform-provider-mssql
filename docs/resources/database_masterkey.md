# mssql_database_masterkey

The `mssql_database_masterkey` resource manages the creation and configuration of database master keys (DMKs) in SQL Server databases.

## Example Usage

```hcl
resource "mssql_database_masterkey" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {}
  }

  database = "example-db"
  password = "strong password"
}
```

### Using a write-only password

With Terraform 1.11 or later, `password_wo` keeps the password out of the plan and state. Increment
`password_wo_version` to regenerate the master key with a new password.

```hcl
resource "random_password" "password" {
  length = 24
}

resource "mssql_database_masterkey" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {}
  }

  database            = "example-db"
  password_wo         = random_password.password.result
  password_wo_version = 1
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database` - (Required) The name of the database to operate on. Changing this forces a new resource to be created.
* `password` - (Optional) The password that is used to encrypt the master key in the database. The value is stored in the Terraform state. Exactly one of `password` and `password_wo` must be specified. Changing this resource property modifies the existing resource.
* `password_wo` - (Optional) The password that is used to encrypt the master key in the database, as a [write-only argument](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments): it is never stored in the plan or state, and can take an ephemeral value. Requires Terraform 1.11 or later. Conflicts with `password`. Must be used together with `password_wo_version`.
* `password_wo_version` - (Optional) Version of the `password_wo` value. Terraform cannot detect a change to `password_wo`, so change this value (for example, increment it) to regenerate the master key with a new password. Required with `password_wo`.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.
* `azure_login` - (Optional) Azure AD login for managing the database resources. The attributes supported in the `azure_login` block is detailed below.
* `azuread_default_chain_auth` - (Optional) Use a chain of strategies for authenticating when managing the database resources. This auth strategy is very similar to how the Azure CLI authenticates. For more information, see [DefaultAzureCredential](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential). The attributes supported in the `azuread_default_chain_auth` block are detailed below.
* `azuread_managed_identity_auth` - (Optional) Use a managed identity for authenticating when managing the database resources. This is mainly useful for specifying a user-assigned managed identity. The attributes supported in the `azuread_managed_identity_auth` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.

The `azure_login` block supports the following arguments:

* `tenant_id` - (Required) The tenant ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_TENANT_ID` environment variable.
* `client_id` - (Required) The client ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_ID` environment variable.
* `client_secret` - (Required) The client secret of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_SECRET` environment variable.

The `azuread_default_chain_auth` block supports the following arguments:

* `use_oidc` - (Optional) Use a federated/OIDC credential (workload identity federation) instead of the default credential chain. Defaults to `false`. Conflicts with `credential_type`. Required environment variables:

  | Variable | Description |
  |---|---|
  | `ARM_TENANT_ID` | Tenant ID of the App Registration |
  | `ARM_CLIENT_ID` | Client ID of the App Registration |
  | `ARM_OIDC_TOKEN` | Signed JWT token (inline). Use this **or** `ARM_OIDC_TOKEN_FILE_PATH`. |
  | `ARM_OIDC_TOKEN_FILE_PATH` | Path to a file containing the signed JWT. The file is re-read on every token refresh. |

* `credential_type` - (Optional) Use a single credential instead of the default credential chain. Conflicts with `use_oidc`. Valid values:

  | Value | go-mssqldb `fedauth` | Notes |
  |---|---|---|
  | `az_cli` | `ActiveDirectoryAzCli` | Uses the identity signed in to the Azure CLI (`az login`). |
  | `azure_developer_cli` | `ActiveDirectoryAzureDeveloperCli` | Uses the identity signed in to the Azure Developer CLI (`azd auth login`). |
  | `azure_pipelines` | `ActiveDirectoryAzurePipelines` | Uses an Azure DevOps service connection (workload identity federation). Works with `TerraformTask@5` (reads `ARM_CLIENT_ID`, `ARM_TENANT_ID`, `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` and `ARM_OIDC_REQUEST_TOKEN`), or with `AzureCLI@2` / `AzurePowerShell@5` (reads `AZURESUBSCRIPTION_CLIENT_ID`, `AZURESUBSCRIPTION_TENANT_ID`, `AZURESUBSCRIPTION_SERVICE_CONNECTION_ID`, plus `SYSTEM_ACCESSTOKEN` mapped from `$(System.AccessToken)`). |
  | `environment` | `ActiveDirectoryEnvironment` | Uses the `AZURE_TENANT_ID`, `AZURE_CLIENT_ID` and `AZURE_CLIENT_SECRET` / `AZURE_CLIENT_CERTIFICATE_PATH` environment variables. |
  | `workload_identity` | `ActiveDirectoryWorkloadIdentity` | Uses Kubernetes workload identity (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_FEDERATED_TOKEN_FILE`). |

  When omitted, `ActiveDirectoryDefault` is used, so an empty `azuread_default_chain_auth {}` block keeps the default credential chain.

The `azuread_managed_identity_auth` block supports the following arguments:

* `user_id` - (Optional) Id of a user-assigned managed identity to assume. Omitting this property instructs the provider to assume a system-assigned managed identity.

-> Only one of `login`, `azure_login`, `azuread_default_chain_auth` and `azuread_managed_identity_auth` can be specified.

## Attribute Reference

The following attributes are exported:

* `principal_id` - The principal id of this database master key.
* `key_name` - The name of this database master key.
* `key_guid` - The guid of this database master key.
* `symmetric_key_id` - The symmetric key id of this database master key.
* `key_length` - The lenth of this database master key.
* `key_algorithm` - The algorithm of this database master key.
* `algorithm_desc` - The algorithm description of this database master key.

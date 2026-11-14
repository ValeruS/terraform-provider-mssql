# mssql_database_role

The `mssql_database_role` resource allows you to create and manage database-level roles in SQL Server.

## Example Usage

### Basic usage

```hcl
resource "mssql_database_role" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }

  database  = "master"
  role_name = "example-role-name"
}
```

### Using AUTHORIZATION

```hcl
resource "mssql_database_role" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }

  database   = "my-database"
  role_name  = "example-role-name"
  owner_name = "example_username"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `role_name` - (Required) The name of the role. Changing this resource property modifies the existing resource.
* `database` - (Optional) The role will be created in this database. Defaults to `master`. Changing this forces a new resource to be created.
* `owner_name` - (Optional) Is the database user or role that is to own the new role. Changing this resource property modifies the existing resource.

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

* `principal_id` - The principal id of this database role.
* `owner_name` - The database user name or role name that is own the role.
* `owning_principal_id` - The database user id or the role id that is own the role.

## Import

Before importing `mssql_database_role`, you must to configure the authentication to your sql server:

1. Using Azure AD authentication, you must set the following environment variables: `MSSQL_TENANT_ID`, `MSSQL_CLIENT_ID` and `MSSQL_CLIENT_SECRET`.
2. Using SQL authentication, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.
3. Using `azuread_default_chain_auth`, add `auth=default_chain` to the import ID, together with `credential_type=<value>` or `use_oidc=true` when the configuration sets them. The `MSSQL_*` environment variables are not used.
4. Using `azuread_managed_identity_auth`, add `auth=managed_identity` to the import ID, together with `user_id=<id>` for a user-assigned identity. The `MSSQL_*` environment variables are not used.

The authentication in the import ID must match the `server` block of the configuration, otherwise the next plan shows a change to the `server` block.

After that you can import the SQL Server database role using the server URL and `role name`, e.g.

```shell
terraform import mssql_database_role.example 'mssql://example-sql-server.database.windows.net/example-db/role/role_name'
```

Using `azuread_default_chain_auth` with `credential_type = "az_cli"`:

```shell
terraform import mssql_database_role.example 'mssql://example-sql-server.database.windows.net/example-db/role/role_name?auth=default_chain&credential_type=az_cli'
```

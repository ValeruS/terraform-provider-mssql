# mssql_database

The `mssql_database` resource allows you to create and manage databases on a SQL Server instance.

-> **Azure SQL Database note:** Although `CREATE DATABASE` can be executed via T-SQL on Azure SQL Database, doing so bypasses Azure Resource Manager entirely. That means you have no control over service tier, compute size (DTUs/vCores), elastic pool membership, backup retention, or geo-replication — all of which must be configured through the Azure API. Use `azurerm_mssql_database` for Azure SQL Database lifecycle management, then use this provider to manage objects *within* those databases (users, roles, permissions, schemas). This resource will return an error if applied against an Azure SQL Database.

## Example Usage

### Basic usage

```hcl
resource "mssql_database" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }

  database_name = "example-db"
}
```

### With explicit collation

```hcl
resource "mssql_database" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }

  database_name = "example-db"
  collation     = "SQL_Latin1_General_CP1_CI_AS"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database_name` - (Required) The name of the database. Changing this resource property modifies the existing resource.
* `collation` - (Optional) The collation of the database, e.g. `SQL_Latin1_General_CP1_CI_AS`. If not specified, the server's default collation is used. Changing this resource property modifies the existing resource.

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

* `use_oidc` - (Optional) When `true`, authenticates using a federated/OIDC credential (workload identity federation) instead of the default credential chain. Credentials are read from the environment variables below. Defaults to `false`.

When `use_oidc = true`, the following environment variables must be set:

| Variable | Description |
|---|---|
| `ARM_TENANT_ID` | Tenant ID of the App Registration |
| `ARM_CLIENT_ID` | Client ID of the App Registration |
| `ARM_OIDC_TOKEN` | Signed JWT token (inline). Use this **or** `ARM_OIDC_TOKEN_FILE_PATH`. |
| `ARM_OIDC_TOKEN_FILE_PATH` | Path to a file containing the signed JWT. The file is re-read on every token refresh. |

The `azuread_managed_identity_auth` block supports the following arguments:

* `user_id` - (Optional) Id of a user-assigned managed identity to assume. Omitting this property instructs the provider to assume a system-assigned managed identity.

-> Only one of `login`, `azure_login`, `azuread_default_chain_auth` and `azuread_managed_identity_auth` can be specified.

## Attribute Reference

The following attributes are exported:

* `database_id` - The internal numeric ID of the database
* `collation` - The collation of the database. Populated from the server when not explicitly set.
* `compatibility_level` - Compatibility level.

## Import

Before importing `mssql_database`, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.

After that you can import the database using the server URL and database name, e.g.

```shell
terraform import mssql_database.example 'sqlserver://localhost:1433/database/example-db'
```

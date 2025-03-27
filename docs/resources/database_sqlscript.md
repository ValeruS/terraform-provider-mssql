# mssql_database_sqlscript

The `mssql_database_sqlscript` resource allows you to execute SQL scripts against a Microsoft SQL Server database. This resource is useful for managing database objects through SQL scripts, such as creating tables, views, stored procedures, or any other database objects.

## Example Usage

```hcl
# Execute an inline SQL script
resource "mssql_database_sqlscript" "create_table" {
  server {
    host = "sqlserver.example.com"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }
  
  database = "MyDatabase"
  script = <<-SQL
    CREATE TABLE Users (
      ID INT PRIMARY KEY,
      Name NVARCHAR(100),
      Email NVARCHAR(255)
    )
  SQL
  
  verify_object = "TABLE Users"
}

# Execute a SQL script from a file
resource "mssql_database_sqlscript" "setup_stored_proc" {
  server {
    host = "sqlserver.example.com"
    login {
      username = "admin"
      password = "password123"
    }
  }
  
  database      = "MyDatabase"
  script_file   = "scripts/stored_procedure.sql"
  verify_object = "PROCEDURE GetUsers"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database` - (Required) The name of the database where the script will be executed.
* `script` - (Required if `script_file` is not set) The SQL script to execute. Conflicts with `script_file`.
* `script_file` - (Required if `script` is not set) Path to a file containing the SQL script to execute. Conflicts with `script`.
* `verify_object` - (Required) Object to verify existence after script execution. Format: 'TYPE NAME' (e.g., 'TABLE Users'). Supported types:
    * `TABLE`
    * `VIEW`
    * `PROCEDURE` or `PROC`
    * `FUNCTION` or `FUNC`
    * `SCHEMA`
    * `TRIGGER` or `TRG`

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Optional) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.
* `azure_login` - (Optional) Azure AD login for managing the database resources. The attributes supported in the `azure_login` block is detailed below.
* `azuread_default_chain_auth` - (Optional) Use a chain of strategies for authenticating when managing the database resources. This auth strategy is very similar to how the Azure CLI authenticates. For more information, see [DefaultAzureCredential](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential). This block has no attributes.
* `azuread_managed_identity_auth` - (Optional) Use a managed identity for authenticating when managing the database resources. This is mainly useful for specifying a user-assigned managed identity. The attributes supported in the `azuread_managed_identity_auth` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.

The `azure_login` block supports the following arguments:

* `tenant_id` - (Required) The tenant ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_TENANT_ID` environment variable.
* `client_id` - (Required) The client ID of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_ID` environment variable.
* `client_secret` - (Required) The client secret of the principal used to login to the SQL Server. Can also be sourced from the `MSSQL_CLIENT_SECRET` environment variable.

The `azuread_managed_identity_auth` block supports the following arguments:

* `user_id` - (Optional) Id of a user-assigned managed identity to assume. Omitting this property instructs the provider to assume a system-assigned managed identity.

-> Only one of `login`, `azure_login`, `azuread_default_chain_auth` and `azuread_managed_identity_auth` can be specified.

## Attribute Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the SQL script resource.

## Notes

1. The script is executed exactly as provided, so ensure proper error handling and idempotency in your SQL scripts.
2. The `verify_object` field is used to check if the script execution was successful by verifying the existence of a specific database object.
3. On deletion, no action is taken as the script has already been executed and the objects created by it remain in the database.
4. The resource supports both inline scripts via the `script` argument and file-based scripts via the `script_file` argument.
5. Either `script` or `script_file` must be specified, but not both. 
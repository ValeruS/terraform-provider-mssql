-> Functionality is limited to Azure SQL Database only for RDBMS or BLOB_STORAGE type

# mssql_azure_external_datasource

The `mssql_azure_external_datasource` resource allows you to create and manage external data sources in Azure SQL Database.

## Example Usage

```hcl
resource "mssql_database_masterkey" "name" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = tenant_id
      client_id     = client_id
      client_secret = client_secret
    }
  }

  database = "dbname"
  password = "strongpassword"
}

resource "mssql_database_credential" "example" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {}
  }
  database        = "example-db"
  credential_name = "example-credential-name"
  identity_name   = "example-identity"
  secret          = "strong secret"
}

resource "mssql_azure_external_datasource" "rdbms" {
  server {
    host = "example-sql-server.database.windows.net"
    azure_login {
      tenant_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_id     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      client_secret = "terriblySecretSecret"
    }
  }
  database             = "example_db"
  data_source_name     = "example_name"
  location             = "remote_server_name.database.windows.net"
  credential_name      = mssql_database_credential.example.credential_name
  type                 = "RDBMS"
  remote_database_name = "remoteDB"
}
```
## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `database` - (Required) The name of the database to operate on. Changing this forces a new resource to be created.
* `data_source_name` - (Required) Specifies the name of the external data source being created. Changing this forces a new resource to be created.
* `location` - (Required) Provides the connectivity protocol and path to the external data source. Changing this resource property modifies the existing resource.
* `credential_name` - (Required) Specifies a database-scoped credential for authenticating to the external data source. Changing this resource property modifies the existing resource.
* `type` - (Required) Specifies the type of the external data source being configured. One of either `RDBMS` or `BLOB_STORAGE` must be specified. Changing this forces a new resource to be created.
* `remote_database_name` - (Optional) The name of the remote database on the server provided using `location`. Configure this argument when the `type` is set to `RDBMS`. Changing this resource property modifies the existing resource.

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

The following attributes are exported:

* `data_source_id` - The id of this data source name.
* `credential_id` - The id of the database scoped credential.

## Import

Before importing `mssql_azure_external_datasource`, you must to configure the authentication to your sql server:

1. Using Azure AD authentication, you must set the following environment variables: `MSSQL_TENANT_ID`, `MSSQL_CLIENT_ID` and `MSSQL_CLIENT_SECRET`.
2. Using SQL authentication, you must set the following environment variables: `MSSQL_USERNAME` and `MSSQL_PASSWORD`.

After that you can import the SQL Server database scoped credential using the server URL and `data source name`, e.g.

```shell
terraform import mssql_azure_external_datasource.example 'mssql://example-sql-server.database.windows.net/example-db/data_source_name'
```

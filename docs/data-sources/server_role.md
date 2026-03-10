# mssql_server_role (Data Source)

The `mssql_server_role` data source obtains information about a server-level role in SQL Server.

## Example Usage

```hcl
data "mssql_server_role" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }
  role_name = "example-role-name"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) Server and login details for the SQL Server. The attributes supported in the `server` block is detailed below.
* `role_name` - (Required) The name of the server role.

The `server` block supports the following arguments:

* `host` - (Required) The host of the SQL Server. Changing this forces a new resource to be created.
* `port` - (Optional) The port of the SQL Server. Defaults to `1433`. Changing this forces a new resource to be created.
* `login` - (Required) SQL Server login for managing the database resources. The attributes supported in the `login` block is detailed below.

The `login` block supports the following arguments:

* `username` - (Required) The username of the SQL Server login. Can also be sourced from the `MSSQL_USERNAME` environment variable.
* `password` - (Required) The password of the SQL Server login. Can also be sourced from the `MSSQL_PASSWORD` environment variable.

## Attribute Reference

The following attributes are exported:

* `principal_id` - The principal id of this server role.
* `owner_name` - The server login name that owns the role.
* `owning_principal_id` - The principal id of the login that owns the role.

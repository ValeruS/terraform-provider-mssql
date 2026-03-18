# Microsoft SQL Server Provider

The SQL Server provider exposes resources used to manage the configuration of resources in a Microsoft SQL Server and an Azure SQL Database. It might also work for other Microsoft SQL Server products like Azure Managed SQL Server, but it has not been tested against these resources.

## Example Usage

```hcl
terraform {
  required_providers {
    mssql = {
      source = "valerus/mssql"
      version = "~> 0.3.5"
    }
  }
}

provider "mssql" {
  debug = "false"
}

resource "mssql_login" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }
  login_name = "testlogin"
  password   = "NotSoS3cret?"
}

resource "mssql_user" "example" {
  server {
    host = "localhost"
    login {
      username = "sa"
      password = "MySuperSecr3t!"
    }
  }
  username   = "testuser"
  login_name = mssql_login.example.login_name
}
```
## Azure DevOps example with federated identity
In the Azure DevOps pipeline, export the OIDC token before the Terraform tasks:

```yaml
- task: AzureCLI@2
  displayName: 'Export OIDC token for MSSQL provider'
  inputs:
    azureSubscription: "${{ variables.serviceConnection }}"
    addSpnToEnvironment: true
    scriptType: bash
    scriptLocation: inlineScript
    inlineScript: |
      echo "##vso[task.setvariable variable=ARM_TENANT_ID]$tenantId"
      echo "##vso[task.setvariable variable=ARM_CLIENT_ID]$servicePrincipalId"
      echo "##vso[task.setvariable variable=ARM_OIDC_TOKEN]$idToken"
```

```hcl
resource "mssql_user" "example" {
  server {
    host = "example.database.windows.net"
    azuread_default_chain_auth {
      use_oidc = true
    }
  }
  database  = "dbName"
  username  = "myuser@example.com"
  object_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  roles     = ["db_datareader"]
}
```

## Argument Reference

The following arguments are supported:

* `debug` - (Optional) Either `false` or `true`. Defaults to `false`. If `true`, the provider will write a debug log to `terraform-provider-mssql.log`.

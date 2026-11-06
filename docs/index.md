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
With an Azure Resource Manager service connection that uses workload identity federation, the provider can request the OIDC token from the pipeline itself, using the same environment variables as the `azurerm` provider:

```yaml
- script: tofu apply -auto-approve
  displayName: 'Tofu apply'
  env:
    ARM_USE_OIDC: true
    ARM_TENANT_ID: $(tenantId)
    ARM_CLIENT_ID: $(clientId)
    ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID: $(serviceConnectionId)
    SYSTEM_ACCESSTOKEN: $(System.AccessToken)
```

The token is refreshed automatically, so long runs do not fail when it expires. `ARM_USE_OIDC` enables `use_oidc` for an empty `azuread_default_chain_auth {}` block, or you can set `use_oidc = true` explicitly.

The same settings can be given in the provider block instead, for example when the `azurerm` provider gets its tenant and client IDs from variables:

```hcl
provider "mssql" {
  tenant_id                        = var.tenant_id
  client_id                        = var.client_id
  use_oidc                         = true
  oidc_azure_service_connection_id = var.service_connection_id
}
```

Alternatively, export a token before the Terraform tasks. This token is not refreshed:

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

The following arguments are used by `azuread_default_chain_auth` when OIDC is enabled. They use the same names and environment variables as the `azurerm` provider:

* `tenant_id` - (Optional) The Tenant ID of the service principal. Can also be sourced from the `ARM_TENANT_ID` environment variable.
* `client_id` - (Optional) The Client ID of the service principal. Can also be sourced from the `ARM_CLIENT_ID` environment variable.
* `use_oidc` - (Optional) When `true`, every `azuread_default_chain_auth` block authenticates with OIDC, as if it set `use_oidc = true`. Can also be sourced from the `ARM_USE_OIDC` environment variable. Defaults to `false`.
* `oidc_token` - (Optional) The OIDC ID token. Can also be sourced from the `ARM_OIDC_TOKEN` environment variable.
* `oidc_token_file_path` - (Optional) The path to a file containing an OIDC ID token. The file is re-read on every token refresh. Can also be sourced from the `ARM_OIDC_TOKEN_FILE_PATH` environment variable.
* `oidc_azure_service_connection_id` - (Optional) The Azure DevOps Pipelines service connection ID. Used when neither `oidc_token` nor `oidc_token_file_path` is set. Can also be sourced from the `ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID` or `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` environment variables.
* `oidc_request_token` - (Optional) The Azure DevOps Pipelines `System.AccessToken`. Can also be sourced from the `ARM_OIDC_REQUEST_TOKEN` or `SYSTEM_ACCESSTOKEN` environment variables.

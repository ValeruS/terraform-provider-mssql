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

## Azure DevOps example with the Azure Pipelines credential

`credential_type = "azure_pipelines"` uses an Azure DevOps service connection with workload identity federation. A new token is requested through the service connection for every new connection, so long runs are not affected by token expiry.

With `TerraformTask@5`, no extra setup is needed: for a workload identity federation service connection the task sets `ARM_CLIENT_ID`, `ARM_TENANT_ID`, `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID` and `ARM_OIDC_REQUEST_TOKEN`, which the provider reads. Leave `environmentAzureRmUseIdTokenGeneration` unset (or `false`).

```yaml
- task: TerraformTask@5
  displayName: 'Terraform apply'
  inputs:
    provider: azurerm
    command: apply
    environmentServiceNameAzureRM: "${{ variables.serviceConnection }}"
```

Alternatively, run Terraform inside an `AzureCLI@2` (or `AzurePowerShell@5`) task, which sets the `AZURESUBSCRIPTION_*` variables, and map `SYSTEM_ACCESSTOKEN` explicitly:

```yaml
- task: AzureCLI@2
  displayName: 'Terraform apply'
  inputs:
    azureSubscription: "${{ variables.serviceConnection }}"
    scriptType: bash
    scriptLocation: inlineScript
    inlineScript: terraform apply -auto-approve
  env:
    SYSTEM_ACCESSTOKEN: $(System.AccessToken)
```

```hcl
resource "mssql_user" "example" {
  server {
    host = "example.database.windows.net"
    azuread_default_chain_auth {
      credential_type = "azure_pipelines"
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

* `debug` - (Optional) Either `false` or `true`. Defaults to `false`. If `true`, the provider will write a debug log to `terraform-provider-mssql.log`. For Azure AD authentication the log also shows which credential was used to acquire the token (for example `DefaultAzureCredential authenticated with AzureCLICredential`); tokens are never logged.

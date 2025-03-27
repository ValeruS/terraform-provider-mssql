package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataAzureExternalDatasource_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataAzureExternalDatasourceDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataAzureExternalDatasource(t, "data_azure_test", "azure", map[string]interface{}{"database": "testdb", "data_source_name": "data_test_datasource", "location": "fakesqlsrv.database.windows.net", "type": "RDBMS", "remote_database_name": "test_db_remote", "credential_name": "data_test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/externaldatasource/data_test_datasource"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "database", "testdb"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "data_source_name", "data_test_datasource"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("data.mssql_azure_external_datasource.data_azure_test", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("data.mssql_azure_external_datasource.data_azure_test", "data_source_id"),
					resource.TestCheckResourceAttrSet("data.mssql_azure_external_datasource.data_azure_test", "credential_id"),
				),
			},
		},
	})
}

func testAccCheckDataAzureExternalDatasource(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_database_masterkey" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				 }
				database = "{{ .database }}"
				password = "{{ .password }}"
			}
			resource "mssql_database_credential" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				credential_name = "{{ .credential_name }}"
				identity_name = "{{ .identity_name }}"
				{{ with .secret }}secret = "{{ . }}"{{ end }}
				depends_on = [mssql_database_masterkey.{{ .name }}]
			}
			resource "mssql_azure_external_datasource" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				data_source_name = "{{ .data_source_name }}"
				location = "{{ .location }}"
				credential_name = "{{ .credential_name }}"
				type = "{{ .type }}"
				remote_database_name = "{{ .remote_database_name }}"
				depends_on = [mssql_database_credential.{{ .name }}]
			}
			data "mssql_azure_external_datasource" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				data_source_name = "{{ .data_source_name }}"
				depends_on = [mssql_azure_external_datasource.{{ .name }}]
			}`

	data["name"] = name
	data["login"] = login
	if login == "fedauth" || login == "msi" || login == "azure" {
		data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
	} else if login == "login" {
		data["host"] = "localhost"
	} else {
		t.Fatalf("login expected to be one of 'login', 'azure', 'msi', 'fedauth', got %s", login)
	}
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

func testAccCheckDataAzureExternalDatasourceDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_azure_external_datasource" {
			continue
		}
		if rs.Type != "mssql_database_credential" {
			continue
		}
		if rs.Type != "mssql_database_masterkey" {
			continue
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		datasourcename := rs.Primary.Attributes["data_source_name"]
		extdatasource, err := connector.GetAzureExternalDatasource(database, datasourcename)
		if extdatasource != nil {
			return fmt.Errorf("external datasource still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAzureExternalDatasource_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckAzureExternalDatasourceDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAzureExternalDatasource(t, "test_az_ext_datasource", "azure", map[string]interface{}{"database": "testdb", "data_source_name": "test_datasource", "location": "fakesqlsrv.database.windows.net", "type": "RDBMS", "remote_database_name": "test_db_remote", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureExternalDatasourceExists("mssql_azure_external_datasource.test_az_ext_datasource"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/test_datasource"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "type", "RDBMS"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "data_source_name", "test_datasource"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.test_az_ext_datasource", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.test_az_ext_datasource", "data_source_id"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.test_az_ext_datasource", "credential_id"),
				),
			},
		},
	})
}

func TestAccAzureExternalDatasource_Azure_Basic_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckAzureExternalDatasourceDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAzureExternalDatasource(t, "update", "azure", map[string]interface{}{"database": "testdb", "data_source_name": "test_datasource", "location": "fakesqlsrv1.database.windows.net", "type": "RDBMS", "remote_database_name": "test_db_remote", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureExternalDatasourceExists("mssql_azure_external_datasource.update"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "type", "RDBMS"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "location", "fakesqlsrv1.database.windows.net"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.update", "data_source_id"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.update", "credential_id"),
				),
			},
			{
				Config: testAccCheckAzureExternalDatasource(t, "update", "azure", map[string]interface{}{"database": "testdb", "data_source_name": "test_datasource", "location": "fakesqlsrv2.database.windows.net", "type": "RDBMS", "remote_database_name": "test_db_remote", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureExternalDatasourceExists("mssql_azure_external_datasource.update"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "type", "RDBMS"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "location", "fakesqlsrv2.database.windows.net"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_azure_external_datasource.update", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.update", "data_source_id"),
					resource.TestCheckResourceAttrSet("mssql_azure_external_datasource.update", "credential_id"),
				),
			},
		},
	})
}

func testAccCheckAzureExternalDatasource(t *testing.T, name string, login string, data map[string]interface{}) string {
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

func testAccCheckAzureExternalDatasourceDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_azure_external_datasource" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		datasourcename := rs.Primary.Attributes["data_source_name"]
		datasource, err := connector.GetAzureExternalDatasource(database, datasourcename)
		if datasource != nil {
			return fmt.Errorf("external datasource still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckAzureExternalDatasourceExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_azure_external_datasource" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_azure_external_datasource", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		database := rs.Primary.Attributes["database"]
		datasourcename := rs.Primary.Attributes["data_source_name"]
		extdatasource, err := connector.GetAzureExternalDatasource(database, datasourcename)
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
		if extdatasource.DataSourceName != datasourcename {
			return fmt.Errorf("expected to be data_source_name %s, got %s", datasourcename, extdatasource.DataSourceName)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "data_source_name":
				actual = extdatasource.DataSourceName
			case "type":
				actual = extdatasource.TypeStr
			default:
				return fmt.Errorf("unknown property %s", check.name)
			}
			if (check.op == "" || check.op == "==") && !equal(check.expected, actual) {
				return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
			}
			if check.op == "!=" && equal(check.expected, actual) {
				return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
			}
		}
		return nil
	}
}

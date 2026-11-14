package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabaseSchemaPermission_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataDatabaseSchemaPermission(t, "data_local", "login", map[string]interface{}{"database": "master", "schema_name": "schema_perm_data", "username": "schema_perm_user_data", "login_name": "schema_perm_login_data", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "EXECUTE"]`}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "id", "sqlserver://localhost:1433/master/schema_permission/schema_perm_data/schema_perm_user_data"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "database", "master"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "schema_name", "schema_perm_data"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "principal_name", "schema_perm_user_data"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_schema_permission.data_local", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_schema_permission.data_local", "permissions.*", "EXECUTE"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_local", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrPair("data.mssql_database_schema_permission.data_local", "principal_id", "mssql_user.data_local", "principal_id"),
					resource.TestCheckResourceAttrPair("data.mssql_database_schema_permission.data_local", "schema_id", "mssql_database_schema.data_local", "schema_id"),
				),
			},
		},
	})
}

func TestAccDataDatabaseSchemaPermission_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataDatabaseSchemaPermission(t, "data_azure", "azure", map[string]interface{}{"database": "testdb", "schema_name": "schema_perm_data_azure", "username": "schema_perm_user_data_azure", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "INSERT"]`}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/schema_permission/schema_perm_data_azure/schema_perm_user_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "database", "testdb"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "schema_name", "schema_perm_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "principal_name", "schema_perm_user_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_schema_permission.data_azure", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_schema_permission.data_azure", "permissions.*", "INSERT"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_schema_permission.data_azure", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrPair("data.mssql_database_schema_permission.data_azure", "principal_id", "mssql_user.data_azure", "principal_id"),
					resource.TestCheckResourceAttrPair("data.mssql_database_schema_permission.data_azure", "schema_id", "mssql_database_schema.data_azure", "schema_id"),
				),
			},
		},
	})
}

func testAccCheckDataDatabaseSchemaPermission(t *testing.T, name string, login string, data map[string]interface{}) string {
	resources := testAccCheckDatabaseSchemaPermission(t, name, login, data)
	text := `data "mssql_database_schema_permission" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database    = "{{ .database }}"
				schema_name = mssql_database_schema_permission.{{ .name }}.schema_name
				principal_name = mssql_database_schema_permission.{{ .name }}.principal_name
				depends_on  = [mssql_database_schema_permission.{{ .name }}]
			}`
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return resources + "\n" + res
}

func testAccCheckDataDatabaseSchemaPermissionDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_schema_permission" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		schemaName := rs.Primary.Attributes["schema_name"]
		principalName := rs.Primary.Attributes["principal_name"]
		permissions, err := connector.GetDatabaseSchemaPermission(database, schemaName, principalName)
		if permissions != nil {
			return fmt.Errorf("schema permissions still exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

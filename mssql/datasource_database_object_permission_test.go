package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabaseObjectPermission_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataDatabaseObjectPermission(t, "data_local", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_data", "username": "obj_perm_user_data", "login_name": "obj_perm_login_data", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "INSERT"]`}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "id", "sqlserver://localhost:1433/master/object_permission/dbo/obj_perm_data/obj_perm_user_data"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "database", "master"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "schema_name", "dbo"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "object_name", "obj_perm_data"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "principal_name", "obj_perm_user_data"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_local", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_local", "permissions.*", "INSERT"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_local", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrPair("data.mssql_database_object_permission.data_local", "principal_id", "mssql_user.data_local", "principal_id"),
					resource.TestCheckResourceAttrPair("data.mssql_database_object_permission.data_local", "object_id", "mssql_database_object_permission.data_local", "object_id"),
				),
			},
		},
	})
}

func TestAccDataDatabaseObjectPermission_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataDatabaseObjectPermission(t, "data_azure", "azure", map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_data_azure", "username": "obj_perm_user_data_azure", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "INSERT"]`}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/object_permission/dbo/obj_perm_data_azure/obj_perm_user_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "database", "testdb"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "object_name", "obj_perm_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "principal_name", "obj_perm_user_data_azure"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_azure", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_azure", "permissions.*", "INSERT"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_azure", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrPair("data.mssql_database_object_permission.data_azure", "principal_id", "mssql_user.data_azure", "principal_id"),
					resource.TestCheckResourceAttrPair("data.mssql_database_object_permission.data_azure", "object_id", "mssql_database_object_permission.data_azure", "object_id"),
				),
			},
		},
	})
}

func TestAccDataDatabaseObjectPermission_Local_Column(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataDatabaseObjectPermission(t, "data_column", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_data_column", "column_name": "name", "username": "obj_perm_user_data_column", "login_name": "obj_perm_login_data_column", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "UPDATE"]`}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_column", "id", "sqlserver://localhost:1433/master/object_permission/dbo/obj_perm_data_column/obj_perm_user_data_column/name"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_column", "column_name", "name"),
					resource.TestCheckResourceAttr("data.mssql_database_object_permission.data_column", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_column", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("data.mssql_database_object_permission.data_column", "permissions.*", "UPDATE"),
				),
			},
		},
	})
}

func testAccCheckDataDatabaseObjectPermission(t *testing.T, name string, login string, data map[string]interface{}) string {
	resources := testAccCheckDatabaseObjectPermission(t, name, login, data)
	text := `data "mssql_database_object_permission" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database       = "{{ .database }}"
				schema_name    = mssql_database_object_permission.{{ .name }}.schema_name
				object_name    = mssql_database_object_permission.{{ .name }}.object_name
				{{ with .column_name }}column_name    = mssql_database_object_permission.{{ $.name }}.column_name{{ end }}
				principal_name = mssql_database_object_permission.{{ .name }}.principal_name
				depends_on     = [mssql_database_object_permission.{{ .name }}]
			}`
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return resources + "\n" + res
}

// testAccCheckDataDatabaseObjectPermissionDestroy checks the principal is gone, then drops the test object, since
// mssql_database_sqlscript does not drop what its script created.
func testAccCheckDataDatabaseObjectPermissionDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_object_permission" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		schemaName := rs.Primary.Attributes["schema_name"]
		objectName := rs.Primary.Attributes["object_name"]
		columnName := rs.Primary.Attributes["column_name"]
		principalName := rs.Primary.Attributes["principal_name"]
		permissions, err := connector.GetDatabaseObjectPermission(database, schemaName, objectName, columnName, principalName)
		if permissions != nil {
			return fmt.Errorf("object permissions still exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		drop := fmt.Sprintf(`DECLARE @type char(2) = (SELECT [type] FROM [sys].[objects] WHERE object_id = OBJECT_ID(N'[%[1]s].[%[2]s]'))
			IF @type = 'U' DROP TABLE [%[1]s].[%[2]s]
			IF @type = 'P' DROP PROCEDURE [%[1]s].[%[2]s]`, schemaName, objectName)
		if err = connector.DataBaseExecuteScript(database, drop); err != nil {
			return fmt.Errorf("unable to drop test object [%s].[%s]: %s", schemaName, objectName, err)
		}
	}
	return nil
}

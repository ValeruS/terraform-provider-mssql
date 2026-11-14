package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseSchemaPermission_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "test_import", "login", map[string]interface{}{"database": "master", "schema_name": "schema_perm_import", "username": "schema_perm_user_import", "login_name": "schema_perm_login_import", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.test_import", "SELECT", "VIEW DEFINITION"),
				),
			},
			{
				ResourceName:      "mssql_database_schema_permission.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database_schema_permission.test_import", "", "azure=false"),
			},
		},
	})
}

func TestAccDatabaseSchemaPermission_Azure_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "schema_name": "schema_perm_azure_import", "username": "schema_perm_user_azure_import", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.test_import", "SELECT", "VIEW DEFINITION"),
				),
			},
			{
				ResourceName:      "mssql_database_schema_permission.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database_schema_permission.test_import", "", "azure=true"),
			},
		},
	})
}

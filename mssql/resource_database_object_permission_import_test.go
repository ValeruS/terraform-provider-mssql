package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseObjectPermission_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "test_import", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_import", "username": "obj_perm_user_import", "login_name": "obj_perm_login_import", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.test_import", "SELECT", "VIEW DEFINITION"),
				),
			},
			{
				ResourceName:      "mssql_database_object_permission.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database_object_permission.test_import", "", "azure=false"),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Azure_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_azure_import", "username": "obj_perm_user_azure_import", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.test_import", "SELECT", "VIEW DEFINITION"),
				),
			},
			{
				ResourceName:      "mssql_database_object_permission.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database_object_permission.test_import", "", "azure=true"),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Local_ColumnImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "test_import", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_column_import", "column_name": "name", "username": "obj_perm_user_column_import", "login_name": "obj_perm_login_column_import", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "UPDATE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.test_import", "SELECT", "UPDATE"),
				),
			},
			{
				ResourceName:      "mssql_database_object_permission.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database_object_permission.test_import", "", "azure=false"),
			},
		},
	})
}

package mssql

import (
	"regexp"
	"testing"

	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseSQLScript_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSQLScriptDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSQLScript(t, "test_import", "login", map[string]interface{}{"database": "master", "sqlscript": base64testScript, "verify_object": "TABLE TestTable"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSQLScriptExists("mssql_database_sqlscript.test_import"),
				),
			},
			{
				ResourceName:      "mssql_database_sqlscript.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"sqlscript"},
				ImportStateIdFunc: testAccImportStateId("mssql_database_sqlscript.test_import", false),
			},
		},
	})
}

func TestAccDatabaseSQLScript_Local_InvalidObjectTypeImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSQLScriptDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSQLScript(t, "test_import_invalid", "login", map[string]interface{}{"database": "master", "sqlscript": base64testScript, "verify_object": "TABLE TestTable"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSQLScriptExists("mssql_database_sqlscript.test_import_invalid"),
				),
			},
			{
				ResourceName:      "mssql_database_sqlscript.test_import_invalid",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"sqlscript"},
				ImportStateId:     "mssql://localhost:1433/master/sqlscript/" + base64.StdEncoding.EncodeToString([]byte("master:TABBLE TestTable")),
				ExpectError:       regexp.MustCompile(`unsupported object type: TABBLE`),
			},
		},
	})
}

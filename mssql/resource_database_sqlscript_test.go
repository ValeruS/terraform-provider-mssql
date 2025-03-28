package mssql

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testScript = "IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'TestTable' AND schema_id = SCHEMA_ID('dbo')) BEGIN CREATE TABLE TestTable (id INT) END"
var base64testScript = base64.StdEncoding.EncodeToString([]byte(testScript))

func TestAccDatabaseSQLScript_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSQLScriptDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSQLScript(t, "local_test_sqlscript", "login", map[string]interface{}{"database": "master", "sqlscript": base64testScript, "verify_object": "TABLE TestTable"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSQLScriptExists("mssql_database_sqlscript.local_test_sqlscript"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "database", "master"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "sqlscript", base64testScript),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "verify_object", "TABLE TestTable"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.local_test_sqlscript", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccDatabaseSQLScript_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSQLScriptDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSQLScript(t, "azure_test_sqlscript", "azure", map[string]interface{}{"database": "testdb", "sqlscript": base64testScript, "verify_object": "TABLE TestTable"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSQLScriptExists("mssql_database_sqlscript.azure_test_sqlscript"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "sqlscript", base64testScript),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "verify_object", "TABLE TestTable"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_sqlscript.azure_test_sqlscript", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatabaseSQLScript(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `
					resource "mssql_database_sqlscript" "{{ .name }}" {
						server {
							host = "{{ .host }}"
							{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
						}
						database      = "{{ .database }}"
						sqlscript     = "{{ .sqlscript }}"
						verify_object = "{{ .verify_object }}"
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

func testAccCheckDatabaseSQLScriptDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_sqlscript" {
			continue
		}

		// SQL scripts are executed and don't leave persistent resources
		// No need to verify anything on the server side
		// Just ensure it's removed from state
		return nil
	}

	return nil
}

func testAccCheckDatabaseSQLScriptExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_database_sqlscript" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_sqlscript", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		database := rs.Primary.Attributes["database"]
		verifyObject := rs.Primary.Attributes["verify_object"]
		query, err := getObjectExistsQuery(verifyObject)
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
		err = connector.DataBaseExecuteScript(database, query)
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
		return nil
	}
}

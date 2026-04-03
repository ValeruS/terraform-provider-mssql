package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataDatabase_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataDatabaseDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataDatabase(t, "local_basic", "login", map[string]interface{}{"database_name": "tf_acc_datasource_db"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "id", "sqlserver://localhost:1433/database/tf_acc_datasource_db"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "database_name", "tf_acc_datasource_db"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_database.local_basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttrSet("data.mssql_database.local_basic", "database_id"),
					resource.TestCheckResourceAttrSet("data.mssql_database.local_basic", "collation"),
					resource.TestCheckResourceAttrSet("data.mssql_database.local_basic", "compatibility_level"),
				),
			},
		},
	})
}

func testAccDataDatabase(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_database" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database_name = "{{ .database_name }}"
			}
			data "mssql_database" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database_name = "{{ .database_name }}"
				depends_on = [mssql_database.{{ .name }}]
			}`

	data["name"] = name
	data["login"] = login
	switch login {
	case "fedauth", "msi", "azure":
		data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
	case "login":
		data["host"] = "localhost"
	default:
		t.Fatalf("login expected to be one of 'login', 'azure', 'msi', 'fedauth', got %s", login)
	}
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

func testAccCheckDataDatabaseDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		databaseName := rs.Primary.Attributes["database_name"]
		db, err := connector.GetDatabase(databaseName)
		if db != nil {
			return fmt.Errorf("database [%s] still exists", databaseName)
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabase_Local_Basic_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabase(t, "local_basic_create", "login", map[string]interface{}{"name": "tf_acc_test_db_create"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("mssql_database.local_basic_create"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "name", "tf_acc_test_db_create"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_database.local_basic_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttrSet("mssql_database.local_basic_create", "database_id"),
					resource.TestCheckResourceAttrSet("mssql_database.local_basic_create", "collation"),
				),
			},
		},
	})
}

func TestAccDatabase_Local_Basic_Create_Collation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabase(t, "local_collation_create", "login", map[string]interface{}{
					"name":      "tf_acc_test_db_collation",
					"collation": "SQL_Latin1_General_CP1_CI_AS",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("mssql_database.local_collation_create",
						Check{"collation", "==", "SQL_Latin1_General_CP1_CI_AS"},
					),
					resource.TestCheckResourceAttr("mssql_database.local_collation_create", "name", "tf_acc_test_db_collation"),
					resource.TestCheckResourceAttr("mssql_database.local_collation_create", "collation", "SQL_Latin1_General_CP1_CI_AS"),
				),
			},
		},
	})
}

func TestAccDatabase_Local_Basic_Update_Collation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabase(t, "local_collation_update", "login", map[string]interface{}{
					"name":      "tf_acc_test_db_coll_update",
					"collation": "SQL_Latin1_General_CP1_CI_AS",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("mssql_database.local_collation_update",
						Check{"collation", "==", "SQL_Latin1_General_CP1_CI_AS"},
					),
					resource.TestCheckResourceAttr("mssql_database.local_collation_update", "collation", "SQL_Latin1_General_CP1_CI_AS"),
				),
			},
			{
				Config: testAccCheckDatabase(t, "local_collation_update", "login", map[string]interface{}{
					"name":      "tf_acc_test_db_coll_update",
					"collation": "SQL_Latin1_General_CP1_CS_AS",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("mssql_database.local_collation_update",
						Check{"collation", "==", "SQL_Latin1_General_CP1_CS_AS"},
					),
					resource.TestCheckResourceAttr("mssql_database.local_collation_update", "collation", "SQL_Latin1_General_CP1_CS_AS"),
				),
			},
		},
	})
}

func TestAccDatabase_Local_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabase(t, "local_import", "login", map[string]interface{}{"name": "tf_acc_test_db_import"}),
			},
			{
				ResourceName:      "mssql_database.local_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_database.local_import", false),
			},
		},
	})
}

func testAccCheckDatabase(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_database" "{{ .label }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				name = "{{ .name }}"
				{{ with .collation }}collation = "{{ . }}"{{ end }}
			}`

	data["label"] = name
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

func testAccCheckDatabaseDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		dbName := rs.Primary.Attributes["name"]
		db, err := connector.GetDatabase(dbName)
		if db != nil {
			return fmt.Errorf("database [%s] still exists", dbName)
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckDatabaseExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_database" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_database", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		dbName := rs.Primary.Attributes["name"]
		db, err := connector.GetDatabase(dbName)
		if err != nil {
			return fmt.Errorf("error fetching database: %s", err)
		}
		if db == nil {
			return fmt.Errorf("database [%s] does not exist", dbName)
		}
		if db.DatabaseName != dbName {
			return fmt.Errorf("expected database name %s, got %s", dbName, db.DatabaseName)
		}

		for _, check := range checks {
			var actual interface{}
			switch check.name {
			case "collation":
				actual = db.Collation
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

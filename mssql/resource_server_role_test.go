package mssql

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServerRole_Local_Basic_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRole(t, "local_test_create", "login", map[string]interface{}{"role_name": "test_role_create"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_create"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "role_name", "test_role_create"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_create", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_server_role.local_test_create", "principal_id"),
				),
			},
		},
	})
}

func TestAccServerRole_Local_Basic_Create_owner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRole(t, "test_create_auth", "login", map[string]interface{}{"role_name": "test_role_auth", "owner_name": "securityadmin"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.test_create_auth"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "role_name", "test_role_auth"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "owner_name", "securityadmin"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role.test_create_auth", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_server_role.test_create_auth", "principal_id"),
				),
			},
		},
	})
}

func TestAccServerRole_Local_Basic_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRole(t, "local_test_update", "login", map[string]interface{}{"role_name": "test_role_pre"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update", Check{"role_name", "==", "test_role_pre"}),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update", "role_name", "test_role_pre"),
				),
			},
			{
				Config: testAccCheckServerRole(t, "local_test_update", "login", map[string]interface{}{"role_name": "test_role_post"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update", Check{"role_name", "==", "test_role_post"}),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update", "role_name", "test_role_post"),
				),
			},
		},
	})
}

func TestAccServerRole_Local_Basic_Update_owner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRole(t, "local_test_update_owner", "login", map[string]interface{}{"role_name": "test_role_owner", "owner_name": "securityadmin"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update_owner", Check{"owner_name", "==", "securityadmin"}),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_owner", "role_name", "test_role_owner"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_owner", "owner_name", "securityadmin"),
				),
			},
			{
				Config: testAccCheckServerRole(t, "local_test_update_owner", "login", map[string]interface{}{"role_name": "test_role_owner", "owner_name": "dbcreator"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update_owner", Check{"owner_name", "==", "dbcreator"}),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_owner", "role_name", "test_role_owner"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_owner", "owner_name", "dbcreator"),
				),
			},
		},
	})
}

func TestAccServerRole_Local_Basic_Remove_owner(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRole(t, "local_test_update_rm_auth", "login", map[string]interface{}{"role_name": "test_role_owner_rm", "owner_name": "securityadmin"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update_rm_auth"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_rm_auth", "role_name", "test_role_owner_rm"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_rm_auth", "owner_name", "securityadmin"),
				),
			},
			{
				Config: testAccCheckServerRole(t, "local_test_update_rm_auth", "login", map[string]interface{}{"role_name": "test_role_owner_rm", "owner_name": ""}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleExists("mssql_server_role.local_test_update_rm_auth"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_rm_auth", "role_name", "test_role_owner_rm"),
					resource.TestCheckResourceAttr("mssql_server_role.local_test_update_rm_auth", "owner_name", "sa"),
				),
			},
		},
	})
}

func TestAccServerRole_Azure_Basic_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckServerRole(t, "azure_test_create", "azure", map[string]interface{}{"role_name": "test_role_create"}),
				ExpectError: regexp.MustCompile("Statement 'CREATE SERVER ROLE' is not supported in this version of SQL Server"),
			},
		},
	})
}

func testAccCheckServerRole(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_server_role" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				role_name = "{{ .role_name }}"
				{{ with .owner_name }}owner_name = "{{ . }}"{{ end }}
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

func testAccCheckServerRoleDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_server_role" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		roleName := rs.Primary.Attributes["role_name"]
		role, err := connector.GetServerRole(roleName)
		if role != nil {
			return fmt.Errorf("role still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckServerRoleExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_server_role" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_server_role", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		roleName := rs.Primary.Attributes["role_name"]
		role, err := connector.GetServerRole(roleName)
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
		if role.RoleName != roleName {
			return fmt.Errorf("expected to be role %s, got %s", roleName, role.RoleName)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "role_name":
				actual = role.RoleName
			case "owner_name":
				actual = role.OwnerName
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

package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataServerRole_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataServerRoleDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataServerRole(t, "data_local_test", "login", map[string]interface{}{"role_name": "data_test_role"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "id", "sqlserver://localhost:1433/role/data_test_role"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "role_name", "data_test_role"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("data.mssql_server_role.data_local_test", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("data.mssql_server_role.data_local_test", "principal_id"),
				),
			},
		},
	})
}

func testAccCheckDataServerRole(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_server_role" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				role_name = "{{ .role_name }}"
			}
			data "mssql_server_role" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				role_name = "{{ .role_name }}"
				depends_on = [mssql_server_role.{{ .name }}]
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

func testAccCheckDataServerRoleDestroy(state *terraform.State) error {
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

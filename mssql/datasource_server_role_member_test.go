package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceServerRoleMember_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataSourceServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceServerRoleMember(t, "local_basic", "login", map[string]interface{}{"role_name": "sysadmin"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "id", "sqlserver://localhost:1433/role_member/sysadmin"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "members.#", "3"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.local_basic", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceServerRoleMember_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDataSourceServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceServerRoleMember(t, "azure_basic", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/role_member/##MS_LoginManager##"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "members.#", "0"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("data.mssql_server_role_member.azure_basic", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDataSourceServerRoleMember(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `data "mssql_server_role_member" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				role_name = "{{ .role_name }}"
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

func testAccCheckDataSourceServerRoleMemberDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_server_role_member" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		roleName := rs.Primary.Attributes["role_name"]
		members, err := connector.GetServerRoleMember(roleName, nil)
		if members != nil {
			continue
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

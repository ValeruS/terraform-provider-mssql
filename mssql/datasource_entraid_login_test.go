package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataEntraIDLogin_Azure_Basic(t *testing.T) {
	clientUser := os.Getenv("TF_ACC_AZURE_USER_CLIENT_USER")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccDataEntraIDLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataEntraIDLogin(t, "basic", "azure", map[string]interface{}{"login_name": clientUser}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "login_name", clientUser),
					resource.TestCheckResourceAttrSet("data.mssql_entraid_login.basic", "principal_id"),
				),
			},
		},
	})
}

func testAccCheckDataEntraIDLogin(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				{{ with .object_id }}object_id = "{{ . }}"{{ end }}
			}
			data "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = mssql_entraid_login.{{ $.name }}.login_name
				depends_on = [mssql_entraid_login.{{ $.name }}]
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

func testAccDataEntraIDLoginDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_entraid_login" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := connector.GetEntraIDLogin(loginName)
		if login != nil {
			return fmt.Errorf("login still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

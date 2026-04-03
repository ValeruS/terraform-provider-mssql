package mssql

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataLogin_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccDataLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataLogin(t, "basic", "login", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_login.basic", "id", "sqlserver://localhost:1433/login/login_basic"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("data.mssql_login.basic", "principal_id"),
					resource.TestCheckResourceAttrSet("data.mssql_login.basic", "sid"),
				),
			},
		},
	})
}

func TestAccDataLogin_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccDataLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataLogin(t, "basic", "azure", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_login.basic", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/login/login_basic"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("data.mssql_login.basic", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("data.mssql_login.basic", "principal_id"),
					resource.TestCheckResourceAttrSet("data.mssql_login.basic", "sid"),
				),
			},
		},
	})
}

func testAccDataLogin(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				password   = "{{ .password }}"
				{{ with .sid }}sid = "{{ . }}"{{ end }}
				{{ with .default_database }}default_database = "{{ . }}"{{ end }}
				{{ with .default_language }}default_language = "{{ . }}"{{ end }}
			}
			data "mssql_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				depends_on = [mssql_login.{{ .name }}]
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

func testAccDataLoginDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_login" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := connector.GetLogin(loginName)
		if login != nil {
			return fmt.Errorf("login still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

package mssql

import (
	"context"
	"testing"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataEntraIDLogin_Mock(t *testing.T) {
	// Create mock connector with test data
	mockConnector := &MockEntraIDLoginConnector{
		login: &model.EntraIDLogin{
			LoginName:       "test_entraid_login",
			ObjectId:        "test-object-id",
			PrincipalID:     1001,
			Sid:            "test-sid",
			DefaultDatabase: "master",
			DefaultLanguage: "us_english",
		},
	}

	// Create test provider with mock connector
	testProvider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"mssql_entraid_login": resourceEntraIDLogin(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"mssql_entraid_login": dataSourceEntraIDLogin(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return mockConnector, nil
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"mssql": func() (*schema.Provider, error) {
				return testProvider, nil
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			return testAccCheckEntraIDLoginDestroyMock(state, mockConnector)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccDataEntraIDLogin(t, "basic", map[string]interface{}{
					"login_name": "test_entraid_login",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "id", "sqlserver://localhost:1433/login/test_entraid_login"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "login_name", "test_entraid_login"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.login.0.username", "test_user"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.login.0.password", "test_password"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "principal_id", "1001"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "sid", "test-sid"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "default_database", "master"),
					resource.TestCheckResourceAttr("data.mssql_entraid_login.basic", "default_language", "us_english"),
				),
			},
		},
	})
}

func testAccDataEntraIDLogin(t *testing.T, name string, data map[string]interface{}) string {
	text := `resource "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					login {
						username = "{{ .username }}"
						password = "{{ .password }}"
					}
				}
				login_name = "{{ .login_name }}"
				object_id = "test-object-id"
			}
			data "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					login {
						username = "{{ .username }}"
						password = "{{ .password }}"
					}
				}
				login_name = "{{ .login_name }}"
				depends_on = [mssql_entraid_login.{{ .name }}]
			}`
	data["name"] = name
	data["host"] = "localhost"
	data["username"] = "test_user"
	data["password"] = "test_password"
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

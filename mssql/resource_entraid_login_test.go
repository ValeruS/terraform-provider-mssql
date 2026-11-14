package mssql

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccEntraIDLogin_Azure_Basic(t *testing.T) {
	clientUser := os.Getenv("TF_ACC_AZURE_USER_CLIENT_USER")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckEntraIDLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "basic", "azure", map[string]interface{}{"login_name": clientUser}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExists("mssql_entraid_login.basic"),
					resource.TestCheckResourceAttr("mssql_entraid_login.basic", "login_name", clientUser),
					resource.TestCheckResourceAttrSet("mssql_entraid_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccEntraIDLogin_Azure_ObjectId(t *testing.T) {
	clientUser := os.Getenv("TF_ACC_AZURE_USER_CLIENT_USER")
	objectId := os.Getenv("TF_ACC_AZURE_USER_CLIENT_OBJECTID")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckEntraIDLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "object_id", "azure", map[string]interface{}{"login_name": clientUser, "object_id": objectId}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExists("mssql_entraid_login.object_id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.object_id", "login_name", clientUser),
					resource.TestCheckResourceAttr("mssql_entraid_login.object_id", "object_id", objectId),
					// For a service principal, the SID is derived from its client ID, not from
					// the object ID used to look it up: the GUID in SQL Server byte order.
					resource.TestCheckResourceAttrWith("mssql_entraid_login.object_id", "sid", func(sid string) error {
						g := strings.ReplaceAll(os.Getenv("TF_ACC_AZURE_USER_CLIENT_ID"), "-", "")
						if len(g) != 32 {
							return fmt.Errorf("invalid TF_ACC_AZURE_USER_CLIENT_ID")
						}
						want := "0x" + g[6:8] + g[4:6] + g[2:4] + g[0:2] + g[10:12] + g[8:10] + g[14:16] + g[12:14] + g[16:]
						if !strings.EqualFold(sid, want) {
							return fmt.Errorf("expected sid %s from the client ID, got %s", want, sid)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccEntraIDLogin_Local_ObjectId_Validate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckEntraIDLogin(t, "object_id_validate", "login", map[string]interface{}{"login_name": "entraid_validate", "object_id": "not-a-guid"}),
				ExpectError: regexp.MustCompile(`expected "object_id" to be a valid UUID`),
			},
		},
	})
}

func testAccCheckEntraIDLogin(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				{{ with .object_id }}object_id = "{{ . }}"{{ end }}
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

func testAccCheckEntraIDLoginDestroy(state *terraform.State) error {
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

func testAccCheckEntraIDLoginExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_entraid_login" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_entraid_login", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := connector.GetEntraIDLogin(loginName)
		if login == nil {
			return fmt.Errorf("login does not exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "default_database":
				actual = login.DefaultDatabase
			case "default_language":
				actual = login.DefaultLanguage
			default:
				return fmt.Errorf("unknown property %s", check.name)
			}
			if (check.op == "" || check.op == "==") && check.expected != actual {
				return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
			}
			if check.op == "!=" && check.expected == actual {
				return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
			}
		}
		return nil
	}
}

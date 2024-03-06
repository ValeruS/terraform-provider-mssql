package mssql

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
  "os"
  "testing"
)

func TestAccAadLogin_Azure_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:          func() { testAccPreCheck(t) },
    ProviderFactories: testAccProviders,
    CheckDestroy:      func(state *terraform.State) error { return testAccCheckAadLoginDestroy(state) },
    Steps: []resource.TestStep{
      {
        Config: testAccCheckAadLogin(t, "basic", true, map[string]interface{}{"login_name": "bob@contoso.com"}),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAadLoginExists("mssql_aad_login.basic"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "login_name", "bob@contoso.com"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "default_database", "master"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "default_language", "us_english"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.#", "1"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.port", "1433"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.azure_login.#", "1"),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
          resource.TestCheckResourceAttr("mssql_aad_login.basic", "server.0.login.#", "0"),
          resource.TestCheckResourceAttrSet("mssql_aad_login.basic", "principal_id"),
          resource.TestCheckResourceAttrSet("mssql_aad_login.basic", "sid"),
        ),
      },
    },
  })
}

func testAccCheckAadLogin(t *testing.T, name string, azure bool, data map[string]interface{}) string {
  text := `resource "mssql_aad_login" "{{ .name }}" {
             server {
               host = "{{ .host }}"
               {{ if .azure }}azure_login {}{{ else }}login {}{{ end }}
             }
             login_name = "{{ .login_name }}"
             {{ with .default_database }}default_database = "{{ . }}"{{ end }}
             {{ with .default_language }}default_language = "{{ . }}"{{ end }}
           }`
  data["name"] = name
  data["azure"] = azure
  if azure {
    data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
  } else {
    data["host"] = "localhost"
  }
  res, err := templateToString(name, text, data)
  if err != nil {
    t.Fatalf("%s", err)
  }
  return res
}

func testAccCheckAadLoginDestroy(state *terraform.State) error {
  for _, rs := range state.RootModule().Resources {
    if rs.Type != "mssql_aad_login" {
      continue
    }

    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetAadLogin(loginName)
    if login != nil {
      return fmt.Errorf("login still exists")
    }
    if err != nil {
      return fmt.Errorf("expected no error, got %s", err)
    }
  }
  return nil
}

func testAccCheckAadLoginExists(resource string, checks ...Check) resource.TestCheckFunc {
  return func(state *terraform.State) error {
    rs, ok := state.RootModule().Resources[resource]
    if !ok {
      return fmt.Errorf("not found: %s", resource)
    }
    if rs.Type != "mssql_aad_login" {
      return fmt.Errorf("expected resource of type %s, got %s", "mssql_aad_login", rs.Type)
    }
    if rs.Primary.ID == "" {
      return fmt.Errorf("no record ID is set")
    }
    connector, err := getTestConnector(rs.Primary.Attributes)
    if err != nil {
      return err
    }

    loginName := rs.Primary.Attributes["login_name"]
    login, err := connector.GetAadLogin(loginName)
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

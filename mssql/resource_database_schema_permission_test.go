package mssql

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseSchemaPermission_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "basic", "login", map[string]interface{}{"database": "master", "schema_name": "schema_perm_basic", "username": "schema_perm_user_basic", "login_name": "schema_perm_login_basic", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.basic", "SELECT", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "id", "sqlserver://localhost:1433/master/schema_permission/schema_perm_basic/schema_perm_user_basic"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "database", "master"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "schema_name", "schema_perm_basic"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "principal_name", "schema_perm_user_basic"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.basic", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.basic", "permissions.*", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrPair("mssql_database_schema_permission.basic", "principal_id", "mssql_user.basic", "principal_id"),
					resource.TestCheckResourceAttrPair("mssql_database_schema_permission.basic", "schema_id", "mssql_database_schema.basic", "schema_id"),
				),
			},
		},
	})
}

func TestAccDatabaseSchemaPermission_Local_Update(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "master", "schema_name": "schema_perm_update", "username": "schema_perm_user_update", "login_name": "schema_perm_login_update", "login_password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "update", "login", data(`["SELECT"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.update", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.update", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "SELECT"),
				),
			},
			// Grant only
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "update", "login", data(`["SELECT", "INSERT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.update", "INSERT", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "INSERT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "UPDATE"),
				),
			},
			// Revoke only
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "update", "login", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.update", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.update", "permissions.#", "2"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "update", "login", data(`["EXECUTE", "DELETE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.update", "DELETE", "EXECUTE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.update", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "EXECUTE"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.update", "permissions.*", "DELETE"),
				),
			},
		},
	})
}

func TestAccDatabaseSchemaPermission_Local_Role(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "role", "login", map[string]interface{}{"database": "master", "schema_name": "schema_perm_role", "username": "schema_perm_user_role", "login_name": "schema_perm_login_role", "login_password": "valueIsH8kd$¡", "role_name": "schema_perm_role_name", "permissions": `["SELECT", "EXECUTE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.role", "EXECUTE", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.role", "id", "sqlserver://localhost:1433/master/schema_permission/schema_perm_role/schema_perm_role_name"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.role", "principal_name", "schema_perm_role_name"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.role", "permissions.#", "2"),
					resource.TestCheckResourceAttrPair("mssql_database_schema_permission.role", "principal_id", "mssql_database_role.role", "principal_id"),
				),
			},
		},
	})
}

func TestAccDatabaseSchemaPermission_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "azure", "azure", map[string]interface{}{"database": "testdb", "schema_name": "schema_perm_azure", "username": "schema_perm_user_azure", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "EXECUTE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.azure", "EXECUTE", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/schema_permission/schema_perm_azure/schema_perm_user_azure"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "schema_name", "schema_perm_azure"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "principal_name", "schema_perm_user_azure"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure", "permissions.*", "EXECUTE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrPair("mssql_database_schema_permission.azure", "principal_id", "mssql_user.azure", "principal_id"),
					resource.TestCheckResourceAttrPair("mssql_database_schema_permission.azure", "schema_id", "mssql_database_schema.azure", "schema_id"),
				),
			},
		},
	})
}

func TestAccDatabaseSchemaPermission_Azure_Update(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "testdb", "schema_name": "schema_perm_azure_update", "username": "schema_perm_user_azure_update", "password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseSchemaPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "azure_update", "azure", data(`["SELECT"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.azure_update", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "principal_name", "schema_perm_user_azure_update"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "server.0.login.#", "0"),
				),
			},
			// Grant only
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "azure_update", "azure", data(`["SELECT", "INSERT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.azure_update", "INSERT", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "INSERT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "UPDATE"),
				),
			},
			// Revoke only
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "azure_update", "azure", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.azure_update", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "UPDATE"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseSchemaPermission(t, "azure_update", "azure", data(`["EXECUTE", "DELETE", "VIEW DEFINITION"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseSchemaPermissionExist("mssql_database_schema_permission.azure_update", "DELETE", "EXECUTE", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_schema_permission.azure_update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "EXECUTE"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr("mssql_database_schema_permission.azure_update", "permissions.*", "VIEW DEFINITION"),
				),
			},
		},
	})
}

func testAccCheckDatabaseSchemaPermission(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `{{ if .login_name }}
			resource "mssql_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				password   = "{{ .login_password }}"
			}
			{{ end }}
			resource "mssql_user" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				username = "{{ .username }}"
				{{ with .password }}password = "{{ . }}"{{ end }}
				{{ with .login_name }}login_name = mssql_login.{{ $.name }}.login_name{{ end }}
			}
			resource "mssql_database_schema" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database    = "{{ .database }}"
				schema_name = "{{ .schema_name }}"
			}
			{{ with .role_name }}
			resource "mssql_database_role" "{{ $.name }}" {
				server {
					host = "{{ $.host }}"
					{{if eq $.login "fedauth"}}azuread_default_chain_auth {}{{ else if eq $.login "msi"}}azuread_managed_identity_auth {}{{ else if eq $.login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database  = "{{ $.database }}"
				role_name = "{{ . }}"
			}
			{{ end }}
			resource "mssql_database_schema_permission" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database    = "{{ .database }}"
				schema_name = mssql_database_schema.{{ .name }}.schema_name
				principal_name = {{ if .role_name }}mssql_database_role.{{ .name }}.role_name{{ else }}mssql_user.{{ .name }}.username{{ end }}
				permissions = {{ .permissions }}
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

func testAccCheckDatabaseSchemaPermissionDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_schema_permission" {
			continue
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		schemaName := rs.Primary.Attributes["schema_name"]
		principalName := rs.Primary.Attributes["principal_name"]

		permissions, err := connector.GetDatabaseSchemaPermission(database, schemaName, principalName)
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
		if permissions != nil && len(permissions.Permissions) > 0 {
			return fmt.Errorf("permissions on schema [%s] for principal [%s] still exist: %v", schemaName, principalName, permissions.Permissions)
		}
	}
	return nil
}

// testAccCheckDatabaseSchemaPermissionExist checks the permissions granted on the server match the expected ones.
func testAccCheckDatabaseSchemaPermissionExist(resource string, expected ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_database_schema_permission" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_schema_permission", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		schemaName := rs.Primary.Attributes["schema_name"]
		principalName := rs.Primary.Attributes["principal_name"]

		permissions, err := connector.GetDatabaseSchemaPermission(database, schemaName, principalName)
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
		if permissions == nil {
			return fmt.Errorf("principal [%s] or schema [%s] does not exist", principalName, schemaName)
		}

		actual := append([]string(nil), permissions.Permissions...)
		sort.Strings(actual)
		want := append([]string(nil), expected...)
		sort.Strings(want)
		if strings.Join(actual, ",") != strings.Join(want, ",") {
			return fmt.Errorf("expected permissions %v, got %v", want, actual)
		}
		return nil
	}
}

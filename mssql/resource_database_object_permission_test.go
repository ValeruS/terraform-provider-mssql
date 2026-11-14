package mssql

import (
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseObjectPermission_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "basic", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_basic", "username": "obj_perm_user_basic", "login_name": "obj_perm_login_basic", "login_password": "valueIsH8kd$¡", "permissions": `["SELECT", "VIEW DEFINITION"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.basic", "SELECT", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "id", "sqlserver://localhost:1433/master/object_permission/dbo/obj_perm_basic/obj_perm_user_basic"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "database", "master"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "schema_name", "dbo"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "object_name", "obj_perm_basic"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "principal_name", "obj_perm_user_basic"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.basic", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.basic", "permissions.*", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrPair("mssql_database_object_permission.basic", "principal_id", "mssql_user.basic", "principal_id"),
					resource.TestCheckResourceAttrSet("mssql_database_object_permission.basic", "object_id"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Local_Procedure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "procedure", "login", map[string]interface{}{"database": "master", "object_type": "PROCEDURE", "object_name": "obj_perm_proc", "username": "obj_perm_user_proc", "login_name": "obj_perm_login_proc", "login_password": "valueIsH8kd$¡", "permissions": `["EXECUTE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.procedure", "EXECUTE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.procedure", "object_name", "obj_perm_proc"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.procedure", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.procedure", "permissions.*", "EXECUTE"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Local_Update(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_update", "username": "obj_perm_user_update", "login_name": "obj_perm_login_update", "login_password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "update", "login", data(`["SELECT"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.update", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.update", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "SELECT"),
				),
			},
			// Grant only
			{
				Config: testAccCheckDatabaseObjectPermission(t, "update", "login", data(`["SELECT", "INSERT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.update", "INSERT", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "INSERT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "UPDATE"),
				),
			},
			// Revoke only
			{
				Config: testAccCheckDatabaseObjectPermission(t, "update", "login", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.update", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.update", "permissions.#", "2"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseObjectPermission(t, "update", "login", data(`["DELETE", "REFERENCES", "VIEW DEFINITION"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.update", "DELETE", "REFERENCES", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "REFERENCES"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.update", "permissions.*", "VIEW DEFINITION"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Local_Role(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "role", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_role", "username": "obj_perm_user_role", "login_name": "obj_perm_login_role", "login_password": "valueIsH8kd$¡", "role_name": "obj_perm_role_name", "permissions": `["SELECT", "INSERT"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.role", "INSERT", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.role", "id", "sqlserver://localhost:1433/master/object_permission/dbo/obj_perm_role/obj_perm_role_name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.role", "principal_name", "obj_perm_role_name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.role", "permissions.#", "2"),
					resource.TestCheckResourceAttrPair("mssql_database_object_permission.role", "principal_id", "mssql_database_role.role", "principal_id"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure", "azure", map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_azure", "username": "obj_perm_user_azure", "password": "valueIsH8kd$¡", "permissions": `["SELECT", "INSERT"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure", "INSERT", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/object_permission/dbo/obj_perm_azure/obj_perm_user_azure"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "schema_name", "dbo"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "object_name", "obj_perm_azure"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "principal_name", "obj_perm_user_azure"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure", "permissions.*", "INSERT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrPair("mssql_database_object_permission.azure", "principal_id", "mssql_user.azure", "principal_id"),
					resource.TestCheckResourceAttrSet("mssql_database_object_permission.azure", "object_id"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Azure_Update(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_azure_update", "username": "obj_perm_user_azure_update", "password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_update", "azure", data(`["SELECT"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_update", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "principal_name", "obj_perm_user_azure_update"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "permissions.#", "1"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure_update", "permissions.*", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "server.0.login.#", "0"),
				),
			},
			// Grant only
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_update", "azure", data(`["SELECT", "INSERT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_update", "INSERT", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "permissions.#", "3"),
				),
			},
			// Revoke only
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_update", "azure", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_update", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "permissions.#", "2"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_update", "azure", data(`["DELETE", "REFERENCES", "VIEW DEFINITION"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_update", "DELETE", "REFERENCES", "VIEW DEFINITION"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_update", "permissions.#", "3"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure_update", "permissions.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure_update", "permissions.*", "REFERENCES"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.azure_update", "permissions.*", "VIEW DEFINITION"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Local_Column(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_column", "column_name": "name", "username": "obj_perm_user_column", "login_name": "obj_perm_login_column", "login_password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "column", "login", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.column", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.column", "id", "sqlserver://localhost:1433/master/object_permission/dbo/obj_perm_column/obj_perm_user_column/name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.column", "object_name", "obj_perm_column"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.column", "column_name", "name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.column", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.column", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.column", "permissions.*", "UPDATE"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseObjectPermission(t, "column", "login", data(`["SELECT", "REFERENCES"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.column", "REFERENCES", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.column", "permissions.#", "2"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.column", "permissions.*", "SELECT"),
					resource.TestCheckTypeSetElemAttr("mssql_database_object_permission.column", "permissions.*", "REFERENCES"),
				),
			},
		},
	})
}

// Permissions on the object and on one of its columns are read separately, so different permissions on each do not
// show as drift.
func TestAccDatabaseObjectPermission_Local_ObjectAndColumn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "mixed", "login", map[string]interface{}{"database": "master", "object_type": "TABLE", "object_name": "obj_perm_mixed", "username": "obj_perm_user_mixed", "login_name": "obj_perm_login_mixed", "login_password": "valueIsH8kd$¡", "permissions": `["INSERT", "DELETE"]`, "column_permissions": `["SELECT", "UPDATE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.mixed", "DELETE", "INSERT"),
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.mixed_column", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.mixed", "column_name", ""),
					resource.TestCheckResourceAttr("mssql_database_object_permission.mixed", "permissions.#", "2"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.mixed_column", "column_name", "name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.mixed_column", "permissions.#", "2"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Azure_ObjectAndColumn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_mixed", "azure", map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_azure_mixed", "username": "obj_perm_user_azure_mixed", "password": "valueIsH8kd$¡", "permissions": `["INSERT", "DELETE"]`, "column_permissions": `["SELECT", "UPDATE"]`}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_mixed", "DELETE", "INSERT"),
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_mixed_column", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed", "column_name", ""),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed", "permissions.#", "2"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed_column", "column_name", "name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed_column", "permissions.#", "2"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_mixed_column", "server.0.azure_login.#", "1"),
				),
			},
		},
	})
}

func TestAccDatabaseObjectPermission_Azure_Column(t *testing.T) {
	data := func(permissions string) map[string]interface{} {
		return map[string]interface{}{"database": "testdb", "object_type": "TABLE", "object_name": "obj_perm_azure_column", "column_name": "name", "username": "obj_perm_user_azure_column", "password": "valueIsH8kd$¡", "permissions": permissions}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseObjectPermissionDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_column", "azure", data(`["SELECT", "UPDATE"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_column", "SELECT", "UPDATE"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_column", "id", "sqlserver://"+os.Getenv("TF_ACC_SQL_SERVER")+":1433/testdb/object_permission/dbo/obj_perm_azure_column/obj_perm_user_azure_column/name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_column", "column_name", "name"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_column", "permissions.#", "2"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_column", "server.0.azure_login.#", "1"),
				),
			},
			// Grant and revoke
			{
				Config: testAccCheckDatabaseObjectPermission(t, "azure_column", "azure", data(`["SELECT", "REFERENCES"]`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseObjectPermissionExist("mssql_database_object_permission.azure_column", "REFERENCES", "SELECT"),
					resource.TestCheckResourceAttr("mssql_database_object_permission.azure_column", "permissions.#", "2"),
				),
			},
		},
	})
}

// testAccCheckDatabaseObjectPermission builds a configuration with a user (or a role), an object in the dbo schema
// created with mssql_database_sqlscript, and the permissions on that object, or on one of its columns when column_name
// is set. object_type is TABLE (with the columns id and name) or PROCEDURE. column_permissions adds a second resource
// with permissions on the name column for the same user.
func testAccCheckDatabaseObjectPermission(t *testing.T, name string, login string, data map[string]interface{}) string {
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
			resource "mssql_database_sqlscript" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database      = "{{ .database }}"
				sqlscript     = "{{ .sqlscript }}"
				verify_object = "{{ .object_type }} dbo.{{ .object_name }}"
			}
			resource "mssql_database_object_permission" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database       = "{{ .database }}"
				schema_name    = "dbo"
				object_name    = "{{ .object_name }}"
				{{ with .column_name }}column_name    = "{{ . }}"{{ end }}
				principal_name = {{ if .role_name }}mssql_database_role.{{ .name }}.role_name{{ else }}mssql_user.{{ .name }}.username{{ end }}
				permissions    = {{ .permissions }}
				depends_on     = [mssql_database_sqlscript.{{ .name }}]
			}
			{{ with .column_permissions }}
			resource "mssql_database_object_permission" "{{ $.name }}_column" {
				server {
					host = "{{ $.host }}"
					{{if eq $.login "fedauth"}}azuread_default_chain_auth {}{{ else if eq $.login "msi"}}azuread_managed_identity_auth {}{{ else if eq $.login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database       = "{{ $.database }}"
				schema_name    = "dbo"
				object_name    = "{{ $.object_name }}"
				column_name    = "name"
				principal_name = mssql_user.{{ $.name }}.username
				permissions    = {{ . }}
				depends_on     = [mssql_database_sqlscript.{{ $.name }}]
			}
			{{ end }}`

	objectName := data["object_name"].(string)
	var script string
	switch data["object_type"] {
	case "TABLE":
		script = fmt.Sprintf("IF OBJECT_ID(N'dbo.%[1]s', N'U') IS NULL CREATE TABLE dbo.%[1]s (id INT, name NVARCHAR(50))", objectName)
	case "PROCEDURE":
		script = fmt.Sprintf("IF OBJECT_ID(N'dbo.%[1]s', N'P') IS NULL EXEC('CREATE PROCEDURE dbo.%[1]s AS SELECT 1')", objectName)
	default:
		t.Fatalf("object_type expected to be one of 'TABLE', 'PROCEDURE', got %v", data["object_type"])
	}
	data["sqlscript"] = base64.StdEncoding.EncodeToString([]byte(script))
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

// testAccCheckDatabaseObjectPermissionDestroy checks the permissions are gone, then drops the test object, since
// mssql_database_sqlscript does not drop what its script created.
func testAccCheckDatabaseObjectPermissionDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_object_permission" {
			continue
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		schemaName := rs.Primary.Attributes["schema_name"]
		objectName := rs.Primary.Attributes["object_name"]
		columnName := rs.Primary.Attributes["column_name"]
		principalName := rs.Primary.Attributes["principal_name"]

		permissions, err := connector.GetDatabaseObjectPermission(database, schemaName, objectName, columnName, principalName)
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
		if permissions != nil && len(permissions.Permissions) > 0 {
			return fmt.Errorf("permissions on object [%s].[%s] column [%s] for principal [%s] still exist: %v", schemaName, objectName, columnName, principalName, permissions.Permissions)
		}

		drop := fmt.Sprintf(`DECLARE @type char(2) = (SELECT [type] FROM [sys].[objects] WHERE object_id = OBJECT_ID(N'[%[1]s].[%[2]s]'))
			IF @type = 'U' DROP TABLE [%[1]s].[%[2]s]
			IF @type = 'P' DROP PROCEDURE [%[1]s].[%[2]s]`, schemaName, objectName)
		if err = connector.DataBaseExecuteScript(database, drop); err != nil {
			return fmt.Errorf("unable to drop test object [%s].[%s]: %s", schemaName, objectName, err)
		}
	}
	return nil
}

// testAccCheckDatabaseObjectPermissionExist checks the permissions granted on the server match the expected ones.
func testAccCheckDatabaseObjectPermissionExist(resource string, expected ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_database_object_permission" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_object_permission", rs.Type)
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
		objectName := rs.Primary.Attributes["object_name"]
		columnName := rs.Primary.Attributes["column_name"]
		principalName := rs.Primary.Attributes["principal_name"]

		permissions, err := connector.GetDatabaseObjectPermission(database, schemaName, objectName, columnName, principalName)
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
		if permissions == nil {
			return fmt.Errorf("principal [%s], object [%s].[%s] or column [%s] does not exist", principalName, schemaName, objectName, columnName)
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

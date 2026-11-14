package mssql

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const testImportId = "sqlserver://example.database.windows.net:1433/testdb/user/app_user"

// clearLoginEnv makes sure the login and azure_login environment variables do not
// influence serverFromId.
func clearLoginEnv(t *testing.T) {
	for _, env := range []string{"MSSQL_USERNAME", "MSSQL_PASSWORD", "MSSQL_TENANT_ID", "MSSQL_CLIENT_ID", "MSSQL_CLIENT_SECRET"} {
		t.Setenv(env, "")
	}
}

func TestServerFromId_Auth(t *testing.T) {
	clearLoginEnv(t)
	// Set login variables to check that auth takes precedence over them.
	t.Setenv("MSSQL_USERNAME", "sa")
	t.Setenv("MSSQL_PASSWORD", "secret")

	tests := []struct {
		name  string
		query string
		block string
		want  map[string]interface{}
	}{
		{
			name:  "default chain",
			query: "?auth=default_chain",
			block: "azuread_default_chain_auth",
			want:  map[string]interface{}{"use_oidc": false, "credential_type": ""},
		},
		{
			name:  "default chain with credential type",
			query: "?auth=default_chain&credential_type=az_cli",
			block: "azuread_default_chain_auth",
			want:  map[string]interface{}{"use_oidc": false, "credential_type": "az_cli"},
		},
		{
			name:  "default chain with oidc",
			query: "?auth=default_chain&use_oidc=true",
			block: "azuread_default_chain_auth",
			want:  map[string]interface{}{"use_oidc": true, "credential_type": ""},
		},
		{
			name:  "system-assigned managed identity",
			query: "?auth=managed_identity",
			block: "azuread_managed_identity_auth",
			want:  map[string]interface{}{"user_id": ""},
		},
		{
			name:  "user-assigned managed identity",
			query: "?auth=managed_identity&user_id=00000000-0000-0000-0000-000000000001",
			block: "azuread_managed_identity_auth",
			want:  map[string]interface{}{"user_id": "00000000-0000-0000-0000-000000000001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _, err := serverFromId(testImportId + tt.query)
			if err != nil {
				t.Fatalf("serverFromId returned error: %v", err)
			}
			if len(server) != 1 {
				t.Fatalf("expected one server block, got %d", len(server))
			}
			s := server[0]
			if s["host"] != "example.database.windows.net" || s["port"] != "1433" {
				t.Errorf("unexpected host/port: %v:%v", s["host"], s["port"])
			}
			for _, other := range []string{"login", "azure_login"} {
				if _, ok := s[other]; ok {
					t.Errorf("%s should not be set with auth", other)
				}
			}
			block, ok := s[tt.block].([]map[string]interface{})
			if !ok || len(block) != 1 {
				t.Fatalf("expected one %s block, got %#v", tt.block, s[tt.block])
			}
			if !reflect.DeepEqual(block[0], tt.want) {
				t.Errorf("%s: got %#v, want %#v", tt.block, block[0], tt.want)
			}

			// The block must be accepted by the server schema, as the importers set it with data.Set.
			data := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				serverProp: {
					Type:     schema.TypeList,
					MaxItems: 1,
					Required: true,
					Elem:     &schema.Resource{Schema: getServerSchema(serverProp)},
				},
			}, map[string]interface{}{})
			if err := data.Set(serverProp, server); err != nil {
				t.Fatalf("data.Set(server) returned error: %v", err)
			}
			if n := len(data.Get(serverProp + ".0." + tt.block).([]interface{})); n != 1 {
				t.Errorf("expected %s in resource data, got %d blocks", tt.block, n)
			}
		})
	}
}

func TestServerFromId_AuthErrors(t *testing.T) {
	clearLoginEnv(t)

	tests := []struct {
		name  string
		query string
		err   string
	}{
		{"unknown auth", "?auth=password", `invalid auth "password"`},
		{"auth with username", "?auth=default_chain&username=sa", "username cannot be used with auth"},
		{"auth with client secret", "?auth=managed_identity&client_secret=x", "client_secret cannot be used with auth"},
		{"auth with azure", "?auth=default_chain&azure=true", "azure cannot be used with auth"},
		{"invalid credential type", "?auth=default_chain&credential_type=password", `invalid credential_type "password"`},
		{"invalid use_oidc", "?auth=default_chain&use_oidc=maybe", `invalid use_oidc "maybe"`},
		{"oidc with credential type", "?auth=default_chain&use_oidc=true&credential_type=az_cli", "use_oidc and credential_type cannot be used together"},
		{"user_id with default chain", "?auth=default_chain&user_id=x", "user_id cannot be used with auth=default_chain"},
		{"credential type with managed identity", "?auth=managed_identity&credential_type=az_cli", "credential_type cannot be used with auth=managed_identity"},
		{"credential type without auth", "?credential_type=az_cli", "credential_type in the import ID requires the auth parameter"},
		{"user_id without auth", "?user_id=x", "user_id in the import ID requires the auth parameter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := serverFromId(testImportId + tt.query)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("error %q does not contain %q", err, tt.err)
			}
		})
	}
}

func TestServerFromId_WithoutAuth(t *testing.T) {
	clearLoginEnv(t)
	t.Setenv("MSSQL_USERNAME", "sa")
	t.Setenv("MSSQL_PASSWORD", "secret")

	server, _, err := serverFromId(testImportId)
	if err != nil {
		t.Fatalf("serverFromId returned error: %v", err)
	}
	login, ok := server[0]["login"].([]map[string]interface{})
	if !ok || len(login) != 1 || login[0]["username"] != "sa" {
		t.Errorf("expected the login block from the environment, got %#v", server[0]["login"])
	}
	if _, ok := server[0]["azuread_default_chain_auth"]; ok {
		t.Error("azuread_default_chain_auth should not be set without auth")
	}
}

package mssql

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/ValeruS/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testImportId = "sqlserver://example.database.windows.net:1433/testdb/user/app_user"

// clearLoginEnv makes sure the login and azure_login environment variables do not
// influence serverFromId.
func clearLoginEnv(t *testing.T) {
	for _, env := range []string{"MSSQL_USERNAME", "MSSQL_PASSWORD", "MSSQL_TENANT_ID", "MSSQL_CLIENT_ID", "MSSQL_CLIENT_SECRET"} {
		t.Setenv(env, "")
	}
}

// testServerFromIdAuth checks that the import ID query sets only the expected
// auth block, and that the server schema accepts it.
func testServerFromIdAuth(t *testing.T, query, block string, want map[string]interface{}) {
	clearLoginEnv(t)
	// Set login variables to check that auth takes precedence over them.
	t.Setenv("MSSQL_USERNAME", "sa")
	t.Setenv("MSSQL_PASSWORD", "secret")

	server, _, err := serverFromId(testImportId + query)
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
	b, ok := s[block].([]map[string]interface{})
	if !ok || len(b) != 1 {
		t.Fatalf("expected one %s block, got %#v", block, s[block])
	}
	if !reflect.DeepEqual(b[0], want) {
		t.Errorf("%s: got %#v, want %#v", block, b[0], want)
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
	if n := len(data.Get(serverProp + ".0." + block).([]interface{})); n != 1 {
		t.Errorf("expected %s in resource data, got %d blocks", block, n)
	}
}

func TestServerFromId_Auth_DefaultChain(t *testing.T) {
	testServerFromIdAuth(t, "?auth=default_chain", "azuread_default_chain_auth",
		map[string]interface{}{"use_oidc": false, "credential_type": ""})
}

func TestServerFromId_Auth_DefaultChainCredentialType(t *testing.T) {
	testServerFromIdAuth(t, "?auth=default_chain&credential_type=az_cli", "azuread_default_chain_auth",
		map[string]interface{}{"use_oidc": false, "credential_type": "az_cli"})
}

func TestServerFromId_Auth_DefaultChainOIDC(t *testing.T) {
	testServerFromIdAuth(t, "?auth=default_chain&use_oidc=true", "azuread_default_chain_auth",
		map[string]interface{}{"use_oidc": true, "credential_type": ""})
}

func TestServerFromId_Auth_ManagedIdentitySystemAssigned(t *testing.T) {
	testServerFromIdAuth(t, "?auth=managed_identity", "azuread_managed_identity_auth",
		map[string]interface{}{"user_id": ""})
}

func TestServerFromId_Auth_ManagedIdentityUserAssigned(t *testing.T) {
	testServerFromIdAuth(t, "?auth=managed_identity&user_id=00000000-0000-0000-0000-000000000001", "azuread_managed_identity_auth",
		map[string]interface{}{"user_id": "00000000-0000-0000-0000-000000000001"})
}

// testServerFromIdAuthError checks that the import ID query is rejected with the
// expected error.
func testServerFromIdAuthError(t *testing.T, query, want string) {
	clearLoginEnv(t)

	_, _, err := serverFromId(testImportId + query)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err, want)
	}
}

func TestServerFromId_AuthError_Unknown(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=password", `invalid auth "password"`)
}

func TestServerFromId_AuthError_Username(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&username=sa", "username cannot be used with auth")
}

func TestServerFromId_AuthError_ClientSecret(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=managed_identity&client_secret=x", "client_secret cannot be used with auth")
}

func TestServerFromId_AuthError_Azure(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&azure=true", "azure cannot be used with auth")
}

func TestServerFromId_AuthError_InvalidCredentialType(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&credential_type=password", `invalid credential_type "password"`)
}

func TestServerFromId_AuthError_InvalidUseOIDC(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&use_oidc=maybe", `invalid use_oidc "maybe"`)
}

func TestServerFromId_AuthError_OIDCWithCredentialType(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&use_oidc=true&credential_type=az_cli", "use_oidc and credential_type cannot be used together")
}

func TestServerFromId_AuthError_UserIdWithDefaultChain(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=default_chain&user_id=x", "user_id cannot be used with auth=default_chain")
}

func TestServerFromId_AuthError_CredentialTypeWithManagedIdentity(t *testing.T) {
	testServerFromIdAuthError(t, "?auth=managed_identity&credential_type=az_cli", "credential_type cannot be used with auth=managed_identity")
}

func TestServerFromId_AuthError_CredentialTypeWithoutAuth(t *testing.T) {
	testServerFromIdAuthError(t, "?credential_type=az_cli", "credential_type in the import ID requires the auth parameter")
}

func TestServerFromId_AuthError_UserIdWithoutAuth(t *testing.T) {
	testServerFromIdAuthError(t, "?user_id=x", "user_id in the import ID requires the auth parameter")
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

// testServerBlockCredentialsChangeIsInPlace checks that a change of the
// credentials in the server block, for example a rotated MSSQL_PASSWORD, is an
// in-place update: it must not replace the resource, and the resource must
// support update so the new credentials can be saved.
func testServerBlockCredentialsChangeIsInPlace(t *testing.T, name string) {
	r, ok := Provider(nil).ResourcesMap[name]
	if !ok {
		t.Fatalf("resource %s not found", name)
	}
	if _, ok := r.Schema[serverProp]; !ok {
		t.Fatalf("resource %s has no server block", name)
	}
	if r.UpdateContext == nil {
		t.Fatal("resource has a server block but no update")
	}

	// Values for attributes whose CustomizeDiff checks their format.
	overrides := map[string]string{
		sqlscriptProp:    "Q1JFQVRFIFRBQkxFIFVzZXJzIChpZCBpbnQp", // base64 of "CREATE TABLE Users (id int)"
		verifyObjectProp: "TABLE Users",
	}

	attributes := map[string]string{"id": "x"}
	config := map[string]interface{}{}
	// Set the required top-level attributes in state and config, so that the
	// diff can be computed. Only the server attribute is checked below.
	for k, s := range r.Schema {
		if k == serverProp || !s.Required {
			continue
		}
		switch s.Type {
		case schema.TypeString:
			value := "value"
			if v, ok := overrides[k]; ok {
				value = v
			}
			attributes[k] = value
			config[k] = value
		case schema.TypeInt:
			attributes[k] = "1"
			config[k] = 1
		case schema.TypeSet, schema.TypeList:
			attributes[k+".#"] = "0"
			config[k] = []interface{}{}
		}
	}
	for k, v := range map[string]string{
		"server.#":                                 "1",
		"server.0.host":                            "example.database.windows.net",
		"server.0.port":                            "1433",
		"server.0.login.#":                         "1",
		"server.0.login.0.username":                "sa",
		"server.0.login.0.password":                "old",
		"server.0.azure_login.#":                   "0",
		"server.0.azuread_default_chain_auth.#":    "0",
		"server.0.azuread_managed_identity_auth.#": "0",
	} {
		attributes[k] = v
	}
	config[serverProp] = []interface{}{map[string]interface{}{
		"host":  "example.database.windows.net",
		"login": []interface{}{map[string]interface{}{"username": "sa", "password": "new"}},
	}}

	state := &terraform.InstanceState{ID: "x", Attributes: attributes}
	diff, err := r.SimpleDiff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	if err != nil {
		t.Fatalf("diff returned error: %v", err)
	}
	attr, ok := diff.Attributes["server.0.login.0.password"]
	if !ok {
		t.Fatal("expected a diff on server.0.login.0.password")
	}
	if attr.RequiresNew {
		t.Error("a change of the server credentials should not replace the resource")
	}
}

func TestServerBlockCredentialsChangeIsInPlace_AzureExternalDatasource(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_azure_external_datasource")
}

func TestServerBlockCredentialsChangeIsInPlace_Database(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabaseCredential(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_credential")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabaseMasterkey(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_masterkey")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabasePermissions(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_permissions")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabaseRole(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_role")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabaseSchema(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_schema")
}

func TestServerBlockCredentialsChangeIsInPlace_DatabaseSqlscript(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_database_sqlscript")
}

func TestServerBlockCredentialsChangeIsInPlace_EntraidLogin(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_entraid_login")
}

func TestServerBlockCredentialsChangeIsInPlace_Login(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_login")
}

func TestServerBlockCredentialsChangeIsInPlace_ServerRole(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_server_role")
}

func TestServerBlockCredentialsChangeIsInPlace_ServerRoleMember(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_server_role_member")
}

func TestServerBlockCredentialsChangeIsInPlace_User(t *testing.T) {
	testServerBlockCredentialsChangeIsInPlace(t, "mssql_user")
}

// testServerBlockAuthChangeUsesNewMethod checks that an update which replaces the
// authentication block named from in the server block with the block named to
// connects with the new block, not with the removed one.
func testServerBlockAuthChangeUsesNewMethod(t *testing.T, from, to string) {
	blocks := map[string]struct {
		state  map[string]string
		config map[string]interface{}
		want   string
	}{
		"login": {
			state:  map[string]string{"server.0.login.#": "1", "server.0.login.0.username": "sa", "server.0.login.0.password": "p"},
			config: map[string]interface{}{"login": []interface{}{map[string]interface{}{"username": "sa", "password": "p"}}},
			want:   "login",
		},
		"azure_login": {
			state:  map[string]string{"server.0.azure_login.#": "1", "server.0.azure_login.0.tenant_id": "t", "server.0.azure_login.0.client_id": "c", "server.0.azure_login.0.client_secret": "s"},
			config: map[string]interface{}{"azure_login": []interface{}{map[string]interface{}{"tenant_id": "t", "client_id": "c", "client_secret": "s"}}},
			want:   "azure_login",
		},
		"azuread_default_chain_auth": {
			state:  map[string]string{"server.0.azuread_default_chain_auth.#": "1", "server.0.azuread_default_chain_auth.0.use_oidc": "false", "server.0.azuread_default_chain_auth.0.credential_type": ""},
			config: map[string]interface{}{"azuread_default_chain_auth": []interface{}{map[string]interface{}{}}},
			want:   "default_chain",
		},
		"azuread_default_chain_auth_az_cli": {
			state:  map[string]string{"server.0.azuread_default_chain_auth.#": "1", "server.0.azuread_default_chain_auth.0.use_oidc": "false", "server.0.azuread_default_chain_auth.0.credential_type": "az_cli"},
			config: map[string]interface{}{"azuread_default_chain_auth": []interface{}{map[string]interface{}{"credential_type": "az_cli"}}},
			want:   "credential_type",
		},
		"azuread_managed_identity_auth": {
			state:  map[string]string{"server.0.azuread_managed_identity_auth.#": "1", "server.0.azuread_managed_identity_auth.0.user_id": ""},
			config: map[string]interface{}{"azuread_managed_identity_auth": []interface{}{map[string]interface{}{}}},
			want:   "managed_identity",
		},
	}
	fromBlock, ok := blocks[from]
	if !ok {
		t.Fatalf("unknown block %s", from)
	}
	toBlock, ok := blocks[to]
	if !ok {
		t.Fatalf("unknown block %s", to)
	}

	var got string
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			serverProp: {Type: schema.TypeList, MaxItems: 1, Required: true, Elem: &schema.Resource{Schema: getServerSchema(serverProp)}},
		},
		CreateContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
		ReadContext:   func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
		DeleteContext: func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics { return nil },
		UpdateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
			c, err := sql.GetFactory().GetConnector(serverProp, d)
			if err != nil {
				return diag.FromErr(err)
			}
			connector := c.(*sql.Connector)
			switch {
			case connector.Login != nil:
				got = "login"
			case connector.AzureLogin != nil:
				got = "azure_login"
			case connector.FedauthMSI != nil:
				got = "managed_identity"
			case connector.FedauthCredential != "":
				got = "credential_type"
			default:
				got = "default_chain"
			}
			return nil
		},
	}

	attributes := map[string]string{"id": "x", "server.#": "1", "server.0.host": "h", "server.0.port": "1433"}
	for k, v := range fromBlock.state {
		attributes[k] = v
	}
	server := map[string]interface{}{"host": "h"}
	for k, v := range toBlock.config {
		server[k] = v
	}
	state := &terraform.InstanceState{ID: "x", Attributes: attributes}
	diff, err := r.SimpleDiff(context.Background(), state, terraform.NewResourceConfigRaw(map[string]interface{}{serverProp: []interface{}{server}}), nil)
	if err != nil {
		t.Fatalf("diff returned error: %v", err)
	}
	if _, diags := r.Apply(context.Background(), state, diff, nil); diags.HasError() {
		t.Fatalf("apply returned error: %v", diags)
	}
	if got != toBlock.want {
		t.Errorf("update connected with %s, want %s", got, toBlock.want)
	}
}

func TestServerBlockAuthChangeUsesNewMethod_LoginToAzureLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "login", "azure_login")
}

func TestServerBlockAuthChangeUsesNewMethod_LoginToDefaultChain(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "login", "azuread_default_chain_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_LoginToDefaultChainAzCli(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "login", "azuread_default_chain_auth_az_cli")
}

func TestServerBlockAuthChangeUsesNewMethod_LoginToManagedIdentity(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "login", "azuread_managed_identity_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_AzureLoginToLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azure_login", "login")
}

func TestServerBlockAuthChangeUsesNewMethod_AzureLoginToDefaultChain(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azure_login", "azuread_default_chain_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_AzureLoginToDefaultChainAzCli(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azure_login", "azuread_default_chain_auth_az_cli")
}

func TestServerBlockAuthChangeUsesNewMethod_AzureLoginToManagedIdentity(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azure_login", "azuread_managed_identity_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainToLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth", "login")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainToAzureLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth", "azure_login")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainToDefaultChainAzCli(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth", "azuread_default_chain_auth_az_cli")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainToManagedIdentity(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth", "azuread_managed_identity_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainAzCliToLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth_az_cli", "login")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainAzCliToAzureLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth_az_cli", "azure_login")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainAzCliToDefaultChain(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth_az_cli", "azuread_default_chain_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_DefaultChainAzCliToManagedIdentity(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_default_chain_auth_az_cli", "azuread_managed_identity_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_ManagedIdentityToLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_managed_identity_auth", "login")
}

func TestServerBlockAuthChangeUsesNewMethod_ManagedIdentityToAzureLogin(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_managed_identity_auth", "azure_login")
}

func TestServerBlockAuthChangeUsesNewMethod_ManagedIdentityToDefaultChain(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_managed_identity_auth", "azuread_default_chain_auth")
}

func TestServerBlockAuthChangeUsesNewMethod_ManagedIdentityToDefaultChainAzCli(t *testing.T) {
	testServerBlockAuthChangeUsesNewMethod(t, "azuread_managed_identity_auth", "azuread_default_chain_auth_az_cli")
}

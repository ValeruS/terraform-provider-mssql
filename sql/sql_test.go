package sql

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// newFederatedConnector is a helper that builds a Connector with a FedauthOIDC
// configured directly — mirrors what GetConnector does at runtime.
func newFederatedConnector(tenantID, clientID, oidcToken, oidcTokenFilePath string) *Connector {
	return &Connector{
		Host:    "test.database.windows.net",
		Port:    "1433",
		Timeout: 30 * time.Second,
		FedauthOIDC: &FedauthOIDC{
			TenantID:          tenantID,
			ClientID:          clientID,
			OIDCToken:         oidcToken,
			OIDCTokenFilePath: oidcTokenFilePath,
		},
	}
}

// ---------------------------------------------------------------------------
// oidcGetAssertion unit tests
// ---------------------------------------------------------------------------

func TestOIDCGetAssertion_Inline(t *testing.T) {
	c := newFederatedConnector("tid", "cid", "my-jwt-token", "")
	got, err := c.oidcGetAssertion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-jwt-token" {
		t.Errorf("expected %q, got %q", "my-jwt-token", got)
	}
}

func TestOIDCGetAssertion_InlineTakesPrecedenceOverFile(t *testing.T) {
	// If both are set, the inline value should win.
	f := writeTokenFile(t, "file-token")
	c := newFederatedConnector("tid", "cid", "inline-token", f)
	got, err := c.oidcGetAssertion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "inline-token" {
		t.Errorf("expected %q, got %q", "inline-token", got)
	}
}

func TestOIDCGetAssertion_File(t *testing.T) {
	f := writeTokenFile(t, "  file-jwt-token\n")
	c := newFederatedConnector("tid", "cid", "", f)
	got, err := c.oidcGetAssertion(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Whitespace should be trimmed.
	if got != "file-jwt-token" {
		t.Errorf("expected %q, got %q", "file-jwt-token", got)
	}
}

func TestOIDCGetAssertion_FileMissing(t *testing.T) {
	c := newFederatedConnector("tid", "cid", "", "/nonexistent/path/token.jwt")
	_, err := c.oidcGetAssertion(context.Background())
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
}

func TestOIDCGetAssertion_NeitherSet(t *testing.T) {
	c := newFederatedConnector("tid", "cid", "", "")
	_, err := c.oidcGetAssertion(context.Background())
	if err == nil {
		t.Fatal("expected an error when neither client_assertion nor client_assertion_file is set")
	}
}

// ---------------------------------------------------------------------------
// GetConnector wiring: azuread_default_chain_auth with use_oidc = true
// ---------------------------------------------------------------------------

func TestGetConnector_DefaultChainOIDC(t *testing.T) {
	auth := model.AzureAuth{TenantID: "env-tenant", ClientID: "env-client", OIDCToken: "env-token"}

	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host": "test.database.windows.net",
				"port": "1433",
				"azuread_default_chain_auth": []interface{}{
					map[string]interface{}{"use_oidc": true},
				},
				"login":                         []interface{}{},
				"azure_login":                   []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	f := new(factory)
	iface, err := f.GetConnector("server", d, auth)
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c, ok := iface.(*Connector)
	if !ok {
		t.Fatalf("expected *Connector, got %T", iface)
	}
	if c.FedauthOIDC == nil {
		t.Fatal("FedauthOIDC should be set when oidc = true")
	}
	if c.FedauthOIDC.TenantID != "env-tenant" {
		t.Errorf("TenantID: got %q, want %q", c.FedauthOIDC.TenantID, "env-tenant")
	}
	if c.FedauthOIDC.ClientID != "env-client" {
		t.Errorf("ClientID: got %q, want %q", c.FedauthOIDC.ClientID, "env-client")
	}
	if c.FedauthOIDC.OIDCToken != "env-token" {
		t.Errorf("OIDCToken: got %q, want %q", c.FedauthOIDC.OIDCToken, "env-token")
	}
	if c.Login != nil {
		t.Error("Login should be nil")
	}
	if c.AzureLogin != nil {
		t.Error("AzureLogin should be nil")
	}
	if c.FedauthMSI != nil {
		t.Error("FedauthMSI should be nil")
	}
}

func TestGetConnector_DefaultChainOIDC_False(t *testing.T) {
	// use_oidc = false → FedauthOIDC should not be set (falls through to ActiveDirectoryDefault).
	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host": "test.database.windows.net",
				"port": "1433",
				"azuread_default_chain_auth": []interface{}{
					map[string]interface{}{"use_oidc": false},
				},
				"login":                         []interface{}{},
				"azure_login":                   []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	f := new(factory)
	iface, err := f.GetConnector("server", d, model.AzureAuth{})
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.FedauthOIDC != nil {
		t.Error("FedauthOIDC should be nil when oidc = false")
	}
}

// ---------------------------------------------------------------------------
// Azure DevOps Pipelines OIDC
// ---------------------------------------------------------------------------

func TestGetConnector_DefaultChainOIDC_AzurePipelines(t *testing.T) {
	// Provider-level use_oidc applies to an empty azuread_default_chain_auth block.
	auth := model.AzureAuth{
		TenantID:            "env-tenant",
		ClientID:            "env-client",
		UseOIDC:             true,
		ServiceConnectionID: "env-service-connection",
		OIDCRequestToken:    "env-access-token",
	}

	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host":                       "test.database.windows.net",
				"port":                       "1433",
				"azuread_default_chain_auth": []interface{}{map[string]interface{}{}},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	iface, err := new(factory).GetConnector("server", d, auth)
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.FedauthOIDC == nil {
		t.Fatal("FedauthOIDC should be set when the provider use_oidc = true")
	}
	if c.FedauthOIDC.ServiceConnectionID != "env-service-connection" {
		t.Errorf("ServiceConnectionID: got %q, want %q", c.FedauthOIDC.ServiceConnectionID, "env-service-connection")
	}
	if c.FedauthOIDC.SystemAccessToken != "env-access-token" {
		t.Errorf("SystemAccessToken: got %q, want %q", c.FedauthOIDC.SystemAccessToken, "env-access-token")
	}
}

func TestOIDCCredential_AzurePipelines(t *testing.T) {
	t.Setenv("SYSTEM_OIDCREQUESTURI", "https://dev.azure.com/org/project/_apis/distributedtask/hubs/build/plans/x/jobs/y/oidctoken")
	c := newFederatedConnector("tid", "cid", "", "")
	c.FedauthOIDC.ServiceConnectionID = "scid"
	c.FedauthOIDC.SystemAccessToken = "access-token"

	cred, err := c.oidcCredential()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cred.(*azidentity.AzurePipelinesCredential); !ok {
		t.Errorf("expected *azidentity.AzurePipelinesCredential, got %T", cred)
	}
}

func TestOIDCCredential_AzurePipelinesMissingAccessToken(t *testing.T) {
	c := newFederatedConnector("tid", "cid", "", "")
	c.FedauthOIDC.ServiceConnectionID = "scid"

	if _, err := c.oidcCredential(); err == nil {
		t.Fatal("expected error when SYSTEM_ACCESSTOKEN is missing")
	}
}

func TestOIDCCredential_AzurePipelinesMissingClientID(t *testing.T) {
	c := newFederatedConnector("tid", "", "", "")
	c.FedauthOIDC.ServiceConnectionID = "scid"
	c.FedauthOIDC.SystemAccessToken = "access-token"

	if _, err := c.oidcCredential(); err == nil {
		t.Fatal("expected error when ARM_CLIENT_ID is missing")
	}
}

func TestOIDCCredential_TokenTakesPrecedenceOverAzurePipelines(t *testing.T) {
	c := newFederatedConnector("tid", "cid", "inline-token", "")
	c.FedauthOIDC.ServiceConnectionID = "scid"
	c.FedauthOIDC.SystemAccessToken = "access-token"

	cred, err := c.oidcCredential()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cred.(*azidentity.ClientAssertionCredential); !ok {
		t.Errorf("expected *azidentity.ClientAssertionCredential, got %T", cred)
	}
}

// ---------------------------------------------------------------------------
// GetConnector wiring: login
// ---------------------------------------------------------------------------

func TestGetConnector_Login(t *testing.T) {
	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host": "localhost",
				"port": "1433",
				"login": []interface{}{
					map[string]interface{}{"username": "sa", "password": "Secret123!"},
				},
				"azure_login":                   []interface{}{},
				"azuread_default_chain_auth":    []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)
	f := new(factory)
	iface, err := f.GetConnector("server", d, model.AzureAuth{})
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.Login == nil {
		t.Fatal("Login should be set")
	}
	if c.Login.Username != "sa" || c.Login.Password != "Secret123!" {
		t.Errorf("Login: got username=%q password=%q, want sa / Secret123!", c.Login.Username, c.Login.Password)
	}
	if c.AzureLogin != nil {
		t.Error("AzureLogin should be nil")
	}
	if c.FedauthOIDC != nil {
		t.Error("FedauthOIDC should be nil")
	}
	if c.FedauthMSI != nil {
		t.Error("FedauthMSI should be nil")
	}
	if c.Host != "localhost" || c.Port != "1433" {
		t.Errorf("Host/Port: got %s:%s", c.Host, c.Port)
	}
}

// ---------------------------------------------------------------------------
// GetConnector wiring: azure_login
// ---------------------------------------------------------------------------

func TestGetConnector_AzureLogin(t *testing.T) {
	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host":  "example.database.windows.net",
				"port":  "1433",
				"login": []interface{}{},
				"azure_login": []interface{}{
					map[string]interface{}{
						"tenant_id":     "tid-123",
						"client_id":     "cid-456",
						"client_secret": "secret",
					},
				},
				"azuread_default_chain_auth":    []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)
	f := new(factory)
	iface, err := f.GetConnector("server", d, model.AzureAuth{})
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.AzureLogin == nil {
		t.Fatal("AzureLogin should be set")
	}
	if c.AzureLogin.TenantID != "tid-123" || c.AzureLogin.ClientID != "cid-456" || c.AzureLogin.ClientSecret != "secret" {
		t.Errorf("AzureLogin: got tenant=%q client=%q secret=%q", c.AzureLogin.TenantID, c.AzureLogin.ClientID, c.AzureLogin.ClientSecret)
	}
	if c.Login != nil {
		t.Error("Login should be nil")
	}
	if c.FedauthOIDC != nil {
		t.Error("FedauthOIDC should be nil")
	}
	if c.FedauthMSI != nil {
		t.Error("FedauthMSI should be nil")
	}
	if c.Host != "example.database.windows.net" {
		t.Errorf("Host: got %q", c.Host)
	}
}

// ---------------------------------------------------------------------------
// GetConnector wiring: azuread_managed_identity_auth
// ---------------------------------------------------------------------------

func TestGetConnector_ManagedIdentity(t *testing.T) {
	res := serverSchemaResource()
	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host":                       "example.database.windows.net",
				"port":                       "1433",
				"login":                      []interface{}{},
				"azure_login":                []interface{}{},
				"azuread_default_chain_auth": []interface{}{},
				"azuread_managed_identity_auth": []interface{}{
					map[string]interface{}{"user_id": "00000000-0000-0000-0000-000000000001"},
				},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)
	f := new(factory)
	iface, err := f.GetConnector("server", d, model.AzureAuth{})
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.FedauthMSI == nil {
		t.Fatal("FedauthMSI should be set")
	}
	if c.FedauthMSI.UserID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("FedauthMSI.UserID: got %q", c.FedauthMSI.UserID)
	}
}

// serverSchemaResource returns a Resource with the standard server block schema used by GetConnector.
func serverSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host":        {Type: schema.TypeString, Required: true},
						"port":        {Type: schema.TypeString, Optional: true, Default: "1433"},
						"login":       {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"username": {Type: schema.TypeString, Optional: true}, "password": {Type: schema.TypeString, Optional: true}}}},
						"azure_login": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"tenant_id": {Type: schema.TypeString, Optional: true}, "client_id": {Type: schema.TypeString, Optional: true}, "client_secret": {Type: schema.TypeString, Optional: true}}}},
						"azuread_default_chain_auth": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"use_oidc": {Type: schema.TypeBool, Optional: true, Default: false},
								},
							},
						},
						"azuread_managed_identity_auth": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"user_id": {Type: schema.TypeString, Optional: true}}}},
					},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func writeTokenFile(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "token.jwt")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}
	return f
}

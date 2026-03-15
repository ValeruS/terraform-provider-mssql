package sql

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	// Set the env vars that GetConnector reads when use_oidc = true.
	t.Setenv("ARM_TENANT_ID", "env-tenant")
	t.Setenv("ARM_CLIENT_ID", "env-client")
	t.Setenv("ARM_OIDC_TOKEN", "env-token")
	t.Setenv("ARM_OIDC_TOKEN_FILE_PATH", "")

	res := &schema.Resource{Schema: map[string]*schema.Schema{
		"server": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host": {Type: schema.TypeString, Required: true},
					"port": {Type: schema.TypeString, Optional: true, Default: "1433"},
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
					// Stubs so schema validation doesn't trip.
					"login":                        {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"username": {Type: schema.TypeString, Optional: true}, "password": {Type: schema.TypeString, Optional: true}}}},
					"azure_login":                  {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"tenant_id": {Type: schema.TypeString, Optional: true}, "client_id": {Type: schema.TypeString, Optional: true}, "client_secret": {Type: schema.TypeString, Optional: true}}}},
					"azuread_managed_identity_auth": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"user_id": {Type: schema.TypeString, Optional: true}}}},
				},
			},
		},
	}}

	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host": "test.database.windows.net",
				"port": "1433",
				"azuread_default_chain_auth": []interface{}{
					map[string]interface{}{"use_oidc": true},
				},
				"login":                        []interface{}{},
				"azure_login":                  []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	f := new(factory)
	iface, err := f.GetConnector("server", d)
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
	res := &schema.Resource{Schema: map[string]*schema.Schema{
		"server": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host": {Type: schema.TypeString, Required: true},
					"port": {Type: schema.TypeString, Optional: true, Default: "1433"},
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
					"login":                        {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"username": {Type: schema.TypeString, Optional: true}, "password": {Type: schema.TypeString, Optional: true}}}},
					"azure_login":                  {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"tenant_id": {Type: schema.TypeString, Optional: true}, "client_id": {Type: schema.TypeString, Optional: true}, "client_secret": {Type: schema.TypeString, Optional: true}}}},
					"azuread_managed_identity_auth": {Type: schema.TypeList, Optional: true, MaxItems: 1, Elem: &schema.Resource{Schema: map[string]*schema.Schema{"user_id": {Type: schema.TypeString, Optional: true}}}},
				},
			},
		},
	}}

	raw := map[string]interface{}{
		"server": []interface{}{
			map[string]interface{}{
				"host": "test.database.windows.net",
				"port": "1433",
				"azuread_default_chain_auth": []interface{}{
					map[string]interface{}{"use_oidc": false},
				},
				"login":                        []interface{}{},
				"azure_login":                  []interface{}{},
				"azuread_managed_identity_auth": []interface{}{},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, res.Schema, raw)

	f := new(factory)
	iface, err := f.GetConnector("server", d)
	if err != nil {
		t.Fatalf("GetConnector returned error: %v", err)
	}

	c := iface.(*Connector)
	if c.FedauthOIDC != nil {
		t.Error("FedauthOIDC should be nil when oidc = false")
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

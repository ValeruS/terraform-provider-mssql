package mssql

import (
	"context"
	"testing"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/ValeruS/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func configureAuth(t *testing.T, raw map[string]interface{}) model.AzureAuth {
	t.Helper()
	p := Provider(sql.GetFactory())
	d := schema.TestResourceDataRaw(t, p.Schema, raw)
	meta, diags := providerConfigure(context.Background(), d, sql.GetFactory())
	if diags.HasError() {
		t.Fatalf("providerConfigure: %v", diags)
	}
	return meta.(mssqlProvider).auth
}

func clearAuthEnv(t *testing.T) {
	for _, name := range []string{
		"ARM_TENANT_ID", "ARM_CLIENT_ID", "ARM_USE_OIDC", "ARM_OIDC_TOKEN", "ARM_OIDC_TOKEN_FILE_PATH",
		"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID",
		"ARM_OIDC_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN",
	} {
		t.Setenv(name, "")
	}
}

func TestProviderConfigure_AzureAuthFromConfig(t *testing.T) {
	clearAuthEnv(t)
	auth := configureAuth(t, map[string]interface{}{
		"tenant_id":                        "cfg-tenant",
		"client_id":                        "cfg-client",
		"use_oidc":                         true,
		"oidc_azure_service_connection_id": "cfg-service-connection",
		"oidc_request_token":               "cfg-access-token",
	})

	want := model.AzureAuth{
		TenantID:            "cfg-tenant",
		ClientID:            "cfg-client",
		UseOIDC:             true,
		ServiceConnectionID: "cfg-service-connection",
		OIDCRequestToken:    "cfg-access-token",
	}
	if auth != want {
		t.Errorf("got %+v, want %+v", auth, want)
	}
}

func TestProviderConfigure_AzureAuthFromEnv(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_TENANT_ID", "env-tenant")
	t.Setenv("ARM_CLIENT_ID", "env-client")
	t.Setenv("ARM_USE_OIDC", "true")
	t.Setenv("ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "env-service-connection")
	t.Setenv("SYSTEM_ACCESSTOKEN", "env-access-token")

	auth := configureAuth(t, map[string]interface{}{})

	want := model.AzureAuth{
		TenantID:            "env-tenant",
		ClientID:            "env-client",
		UseOIDC:             true,
		ServiceConnectionID: "env-service-connection",
		OIDCRequestToken:    "env-access-token",
	}
	if auth != want {
		t.Errorf("got %+v, want %+v", auth, want)
	}
}

func TestProviderConfigure_AzureAuthConfigOverridesEnv(t *testing.T) {
	clearAuthEnv(t)
	t.Setenv("ARM_TENANT_ID", "env-tenant")
	t.Setenv("ARM_CLIENT_ID", "env-client")

	auth := configureAuth(t, map[string]interface{}{
		"tenant_id": "cfg-tenant",
		"client_id": "cfg-client",
	})

	if auth.TenantID != "cfg-tenant" || auth.ClientID != "cfg-client" {
		t.Errorf("config should override env, got tenant %q client %q", auth.TenantID, auth.ClientID)
	}
}

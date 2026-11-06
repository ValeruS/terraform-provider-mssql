package model

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// AzureAuth holds the provider-level Azure settings used by azuread_default_chain_auth.
type AzureAuth struct {
	TenantID            string
	ClientID            string
	UseOIDC             bool
	OIDCToken           string
	OIDCTokenFilePath   string
	ServiceConnectionID string
	OIDCRequestToken    string
}

type ConnectorFactory interface {
	GetConnector(prefix string, data *schema.ResourceData, auth AzureAuth) (interface{}, error)
}

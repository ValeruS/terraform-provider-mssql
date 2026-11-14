package mssql

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const DefaultPort = "1433"

// Values of the auth query parameter of an import ID.
const (
	importAuthDefaultChain    = "default_chain"
	importAuthManagedIdentity = "managed_identity"
)

// defaultChainCredentialTypes are the valid values of azuread_default_chain_auth.credential_type.
var defaultChainCredentialTypes = []string{
	"az_cli",
	"azure_developer_cli",
	"azure_pipelines",
	"environment",
	"workload_identity",
}

func getServerSchema(prefix string) map[string]*schema.Schema {
	if len(prefix) > 0 {
		prefix = prefix + ".0."
	}
	var LoginMethods = []string{
		prefix + "login",
		prefix + "azure_login",
		prefix + "azuread_default_chain_auth",
		prefix + "azuread_managed_identity_auth",
	}
	return map[string]*schema.Schema{
		"host": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return strings.EqualFold(old, new)
			},
		},
		"port": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  DefaultPort,
		},
		"login": {
			Type:         schema.TypeList,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: LoginMethods,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"username": {
						Type:        schema.TypeString,
						Required:    true,
						DefaultFunc: schema.EnvDefaultFunc("MSSQL_USERNAME", nil),
					},
					"password": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						DefaultFunc: schema.EnvDefaultFunc("MSSQL_PASSWORD", nil),
					},
				},
			},
		},
		"azure_login": {
			Type:         schema.TypeList,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: LoginMethods,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"tenant_id": {
						Type:        schema.TypeString,
						Required:    true,
						DefaultFunc: schema.EnvDefaultFunc("MSSQL_TENANT_ID", nil),
					},
					"client_id": {
						Type:        schema.TypeString,
						Required:    true,
						DefaultFunc: schema.EnvDefaultFunc("MSSQL_CLIENT_ID", nil),
					},
					"client_secret": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						DefaultFunc: schema.EnvDefaultFunc("MSSQL_CLIENT_SECRET", nil),
					},
				},
			},
		},
		"azuread_default_chain_auth": {
			Type:         schema.TypeList,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: LoginMethods,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"use_oidc": {
						Type:          schema.TypeBool,
						Optional:      true,
						Default:       false,
						ConflictsWith: []string{prefix + "azuread_default_chain_auth.0.credential_type"},
					},
					"credential_type": {
						Type:          schema.TypeString,
						Optional:      true,
						ConflictsWith: []string{prefix + "azuread_default_chain_auth.0.use_oidc"},
						ValidateFunc:  validation.StringInSlice(defaultChainCredentialTypes, false),
					},
				},
			},
		},
		"azuread_managed_identity_auth": {
			Type:         schema.TypeList,
			MaxItems:     1,
			Optional:     true,
			ExactlyOneOf: LoginMethods,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

func serverFromId(id string) ([]map[string]interface{}, *url.URL, error) {
	u, err := url.Parse(id)
	if err != nil {
		return nil, nil, err
	}

	if u.Scheme != "sqlserver" && u.Scheme != "mssql" {
		return nil, nil, errors.New("invalid schema in ID")
	}

	host := u.Host
	port := DefaultPort

	if strings.ContainsRune(host, ':') {
		var err error
		if host, port, err = net.SplitHostPort(u.Host); err != nil {
			return nil, nil, err
		}
	}

	values := u.Query()

	if values.Has("auth") {
		server, err := serverFromAuth(values)
		if err != nil {
			return nil, nil, err
		}
		server["host"] = host
		server["port"] = port
		return []map[string]interface{}{server}, u, nil
	}
	for _, key := range []string{"credential_type", "use_oidc", "user_id"} {
		if values.Has(key) {
			return nil, nil, fmt.Errorf("%s in the import ID requires the auth parameter", key)
		}
	}

	login, loginInValues := getLogin(values)
	azureLogin, azureInValues := getAzureLogin(values)
	if login == nil && azureLogin == nil {
		return nil, nil, errors.New("neither login nor azure login specified")
	}
	if loginInValues && azureInValues {
		return nil, nil, errors.New("both login and azure login specified in resource")
	}
	if login != nil && azureLogin != nil {
		// prefer azure login
		azure := true
		if v, ok := values["azure"]; ok {
			azure = len(v) == 0 || strings.ToLower(v[0]) == "true"
		}
		if azure {
			login = nil
		} else {
			azureLogin = nil
		}
	}

	return []map[string]interface{}{{
		"host":        host,
		"port":        port,
		"login":       login,
		"azure_login": azureLogin,
	}}, u, nil
}

// serverFromAuth builds the authentication block of the server block from the auth
// parameter of an import ID. The import ID is the only input of an importer, so it
// must describe the same block as the configuration, or the next plan shows a diff.
func serverFromAuth(values url.Values) (map[string]interface{}, error) {
	for _, key := range []string{"username", "password", "tenant_id", "client_id", "client_secret", "azure"} {
		if values.Has(key) {
			return nil, fmt.Errorf("%s cannot be used with auth in the import ID", key)
		}
	}

	switch auth := values.Get("auth"); auth {
	case importAuthDefaultChain:
		if values.Has("user_id") {
			return nil, fmt.Errorf("user_id cannot be used with auth=%s in the import ID", auth)
		}
		credentialType := values.Get("credential_type")
		if values.Has("credential_type") && !slices.Contains(defaultChainCredentialTypes, credentialType) {
			return nil, fmt.Errorf("invalid credential_type %q in the import ID, expected one of %s", credentialType, strings.Join(defaultChainCredentialTypes, ", "))
		}
		useOidc := false
		if values.Has("use_oidc") {
			var err error
			if useOidc, err = strconv.ParseBool(values.Get("use_oidc")); err != nil {
				return nil, fmt.Errorf("invalid use_oidc %q in the import ID", values.Get("use_oidc"))
			}
		}
		if useOidc && credentialType != "" {
			return nil, errors.New("use_oidc and credential_type cannot be used together in the import ID")
		}
		return map[string]interface{}{
			"azuread_default_chain_auth": []map[string]interface{}{{
				"use_oidc":        useOidc,
				"credential_type": credentialType,
			}},
		}, nil
	case importAuthManagedIdentity:
		for _, key := range []string{"credential_type", "use_oidc"} {
			if values.Has(key) {
				return nil, fmt.Errorf("%s cannot be used with auth=%s in the import ID", key, auth)
			}
		}
		return map[string]interface{}{
			"azuread_managed_identity_auth": []map[string]interface{}{{
				"user_id": values.Get("user_id"),
			}},
		}, nil
	default:
		return nil, fmt.Errorf("invalid auth %q in the import ID, expected %s or %s", auth, importAuthDefaultChain, importAuthManagedIdentity)
	}
}

func getLogin(values url.Values) ([]map[string]interface{}, bool) {
	var inValues bool

	username := values.Get("username")
	if username == "" {
		username = os.Getenv("MSSQL_USERNAME")
	} else {
		inValues = true
	}

	password := values.Get("password")
	if password == "" {
		password = os.Getenv("MSSQL_PASSWORD")
	} else {
		inValues = true
	}

	if username == "" || password == "" {
		return nil, false
	}

	return []map[string]interface{}{{
		"username": username,
		"password": password,
	}}, inValues
}

func getAzureLogin(values url.Values) ([]map[string]interface{}, bool) {
	var inValues bool

	tenantId := values.Get("tenant_id")
	if tenantId == "" {
		tenantId = os.Getenv("MSSQL_TENANT_ID")
	} else {
		inValues = true
	}

	clientId := values.Get("client_id")
	if clientId == "" {
		clientId = os.Getenv("MSSQL_CLIENT_ID")
	} else {
		inValues = true
	}

	clientSecret := values.Get("client_secret")
	if clientSecret == "" {
		clientSecret = os.Getenv("MSSQL_CLIENT_SECRET")
	} else {
		inValues = true
	}

	if tenantId == "" || clientId == "" || clientSecret == "" {
		return nil, false
	}

	return []map[string]interface{}{{
		"tenant_id":     tenantId,
		"client_id":     clientId,
		"client_secret": clientSecret,
	}}, inValues
}

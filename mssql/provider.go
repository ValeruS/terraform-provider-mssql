package mssql

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/ValeruS/terraform-provider-mssql/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type mssqlProvider struct {
	factory model.ConnectorFactory
	logger  *zerolog.Logger
	auth    model.AzureAuth
}

const (
	providerLogFile = "terraform-provider-mssql.log"
)

var (
	defaultTimeout = schema.DefaultTimeout(30 * time.Second)
)

func New(version, commit string) func() *schema.Provider {
	return func() *schema.Provider {
		return Provider(sql.GetFactory())
	}
}

func Provider(factory model.ConnectorFactory) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"debug": {
				Type:        schema.TypeBool,
				Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
				Optional:    true,
				Default:     false,
			},
			"tenant_id": {
				Type:        schema.TypeString,
				Description: "The Tenant ID used by `azuread_default_chain_auth` with `use_oidc`.",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},
			"client_id": {
				Type:        schema.TypeString,
				Description: "The Client ID used by `azuread_default_chain_auth` with `use_oidc`.",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},
			"use_oidc": {
				Type:        schema.TypeBool,
				Description: "Use OIDC for every `azuread_default_chain_auth` block.",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_OIDC", false),
			},
			"oidc_token": {
				Type:        schema.TypeString,
				Description: "The OIDC ID token for workload identity federation.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_OIDC_TOKEN", ""),
			},
			"oidc_token_file_path": {
				Type:        schema.TypeString,
				Description: "The path to a file containing an OIDC ID token for workload identity federation.",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_OIDC_TOKEN_FILE_PATH", ""),
			},
			"oidc_azure_service_connection_id": {
				Type:        schema.TypeString,
				Description: "The Azure DevOps Pipelines service connection ID used to request an OIDC token.",
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID"}, ""),
			},
			"oidc_request_token": {
				Type:        schema.TypeString,
				Description: "The Azure DevOps Pipelines system access token used to request an OIDC token.",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN"}, ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"mssql_login":                     resourceLogin(),
			"mssql_user":                      resourceUser(),
			"mssql_database_permissions":      resourceDatabasePermissions(),
			"mssql_database_role":             resourceDatabaseRole(),
			"mssql_database_schema":           resourceDatabaseSchema(),
			"mssql_database_masterkey":        resourceDatabaseMasterkey(),
			"mssql_database_credential":       resourceDatabaseCredential(),
			"mssql_azure_external_datasource": resourceAzureExternalDatasource(),
			"mssql_database_sqlscript":        resourceDatabaseSQLScript(),
			"mssql_entraid_login":             resourceEntraIDLogin(),
			"mssql_server_role":               resourceServerRole(),
			"mssql_server_role_member":        resourceServerRoleMember(),
			"mssql_database":                  resourceDatabase(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"mssql_login":                     dataSourceLogin(),
			"mssql_user":                      dataSourceUser(),
			"mssql_database_permissions":      dataSourceDatabasePermissions(),
			"mssql_database_role":             dataSourceDatabaseRole(),
			"mssql_database_schema":           dataSourceDatabaseSchema(),
			"mssql_database_credential":       datasourceDatabaseCredential(),
			"mssql_azure_external_datasource": datasourceAzureExternalDatasource(),
			"mssql_entraid_login":             dataSourceEntraIDLogin(),
			"mssql_server_role":               dataSourceServerRole(),
			"mssql_server_role_member":        dataSourceServerRoleMember(),
			"mssql_database":                  dataSourceDatabase(),
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return providerConfigure(ctx, data, factory)
		},
	}
}

func providerConfigure(ctx context.Context, data *schema.ResourceData, factory model.ConnectorFactory) (model.Provider, diag.Diagnostics) {
	isDebug := data.Get("debug").(bool)
	logger := newLogger(isDebug)

	logger.Info().Msg("Created provider")

	auth := model.AzureAuth{
		TenantID:            data.Get("tenant_id").(string),
		ClientID:            data.Get("client_id").(string),
		UseOIDC:             data.Get("use_oidc").(bool),
		OIDCToken:           data.Get("oidc_token").(string),
		OIDCTokenFilePath:   data.Get("oidc_token_file_path").(string),
		ServiceConnectionID: data.Get("oidc_azure_service_connection_id").(string),
		OIDCRequestToken:    data.Get("oidc_request_token").(string),
	}

	return mssqlProvider{factory: factory, logger: logger, auth: auth}, nil
}

func (p mssqlProvider) GetConnector(prefix string, data *schema.ResourceData) (interface{}, error) {
	return p.factory.GetConnector(prefix, data, p.auth)
}

func (p mssqlProvider) ResourceLogger(resource, function string) zerolog.Logger {
	return p.logger.With().Str("resource", resource).Str("func", function).Logger()
}

func (p mssqlProvider) DataSourceLogger(datasource, function string) zerolog.Logger {
	return p.logger.With().Str("datasource", datasource).Str("func", function).Logger()
}

func newLogger(isDebug bool) *zerolog.Logger {
	var writer io.Writer = nil
	logLevel := zerolog.Disabled
	if isDebug {
		f, err := os.OpenFile(providerLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Err(err).Msg("error opening file")
		}
		writer = f
		logLevel = zerolog.DebugLevel
	}
	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()
	return &logger
}

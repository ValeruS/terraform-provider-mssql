package mssql

import (
	"context"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func datasourceAzureExternalDatasource() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceAzureExternalDatasourceRead,
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			databaseProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			datasourcenameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			datasourceIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			locationProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			credentialNameProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			typeStrProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			rdatabasenameProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			credentialIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: defaultTimeout,
		},
	}
}

func datasourceAzureExternalDatasourceRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	mssqlversion, err := connector.GetMSSQLVersion(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "unable to get MSSQL version"))
	}
	if !strings.Contains(mssqlversion, "Microsoft SQL Azure") {
		return diag.Errorf("Error: The database is not an Azure SQL Database.")
	}

	datasource, err := connector.GetAzureExternalDatasource(ctx, database, datasourcename)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read external data source [%s] on database [%s]", datasourcename, database))
	}
	if datasource == nil {
		logger.Info().Msgf("No external data source [%s] found on database [%s]", datasourcename, database)
		data.SetId("")
	} else {
		if err = data.Set(datasourcenameProp, datasource.DataSourceName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(datasourceIdProp, datasource.DataSourceId); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(locationProp, datasource.Location); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(typeStrProp, datasource.TypeStr); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(credentialNameProp, datasource.CredentialName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(credentialIdProp, datasource.CredentialId); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(rdatabasenameProp, datasource.RDatabaseName); err != nil {
			return diag.FromErr(err)
		}
		data.SetId(getAzureExternalDatasourceID(data))
	}

	return nil
}

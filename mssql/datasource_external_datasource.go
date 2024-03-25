package mssql

import (
	"context"

	"github.com/betr-io/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func datasourceExternalDatasource() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceExternalDatasourceRead,
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validate.SQLIdentifierName,
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
			typedescProp: {
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
			Default: defaultTimeout,
		},
	}
}

func datasourceExternalDatasourceRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "externaldatasource", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	datasource, err := connector.GetExternalDatasource(ctx, database, datasourcename)
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
		if err = data.Set(typedescProp, datasource.TypeDesc); err != nil {
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
		data.SetId(getDatabaseCredentialID(data))
	}

	return nil
}

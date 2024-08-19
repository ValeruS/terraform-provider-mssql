package mssql

import (
	"context"
	"strings"

	"github.com/betr-io/terraform-provider-mssql/mssql/model"
	"github.com/betr-io/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

const datasourcenameProp = "data_source_name"
const datasourceIdProp = "data_source_id"
const locationProp = "location"
const typedescProp = "type"
const rdatabasenameProp = "remote_database_name"

func resourceAzureExternalDatasource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAzureExternalDatasourceCreate,
		ReadContext:   resourceAzureExternalDatasourceRead,
		UpdateContext: resourceAzureExternalDatasourceUpdate,
		DeleteContext: resourceAzureExternalDatasourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAzureExternalDatasourceImport,
		},
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
				Required: true,
				ForceNew: true,
				ValidateFunc: validate.SQLIdentifier,
			},
			datasourceIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			locationProp: {
				Type:     schema.TypeString,
				Required: true,
			},
			credentialNameProp: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validate.SQLIdentifier,
			},
			typedescProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validate.SQLAzureExternalDatasourceType,
			},
			rdatabasenameProp: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validate.SQLIdentifier,
			},
			credentialIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: defaultTimeout,
			Read: defaultTimeout,
			Update: defaultTimeout,
			Delete: defaultTimeout,
		},
	}
}

type AzureExternalDatasourceConnector interface {
	CreateAzureExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, typedesc, rdatabasename string) error
	GetAzureExternalDatasource(ctx context.Context, database, datasourcename string) (*model.AzureExternalDatasource, error)
	UpdateAzureExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, rdatabasename string) error
	DeleteAzureExternalDatasource(ctx context.Context, database, datasourcename string) error
}

func resourceAzureExternalDatasourceCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "create")
	logger.Debug().Msgf("Create %s", getAzureExternalDatasourceID(data))

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)
	location := data.Get(locationProp).(string)
	credentialname := data.Get(credentialNameProp).(string)
	typedesc := data.Get(typedescProp).(string)
	rdatabasename := data.Get(rdatabasenameProp).(string)

	if (rdatabasename == "") && (typedesc == "RDBMS") {
		return diag.Errorf(rdatabasenameProp + " cannot be empty")
	}

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateAzureExternalDatasource(ctx, database, datasourcename, location, credentialname, typedesc, rdatabasename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create external data source [%s] on database [%s]", datasourcename, database))
	}

	data.SetId(getAzureExternalDatasourceID(data))

	logger.Info().Msgf("created external data source [%s] on database [%s]", datasourcename, database)

	return resourceAzureExternalDatasourceRead(ctx, data, meta)
}

func resourceAzureExternalDatasourceRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	extdatasource, err := connector.GetAzureExternalDatasource(ctx, database, datasourcename)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read external data source [%s] on database [%s]", datasourcename, database))
	}
	if extdatasource == nil {
		logger.Info().Msgf("No external data source [%s] found on database [%s]", datasourcename, database)
		data.SetId("")
	} else {
		if err = data.Set(datasourcenameProp, extdatasource.DataSourceName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(datasourceIdProp, extdatasource.DataSourceId); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(locationProp, extdatasource.Location); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(typedescProp, extdatasource.TypeDesc); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(credentialNameProp, extdatasource.CredentialName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(credentialIdProp, extdatasource.CredentialId); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(rdatabasenameProp, extdatasource.RDatabaseName); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceAzureExternalDatasourceUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "update")
	logger.Debug().Msgf("Update %s", getDatabaseCredentialID(data))

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)
	location := data.Get(locationProp).(string)
	credentialname := data.Get(credentialNameProp).(string)
	rdatabasename := data.Get(rdatabasenameProp).(string)

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.UpdateAzureExternalDatasource(ctx, database, datasourcename, location, credentialname, rdatabasename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to update external data source [%s] on database [%s]", datasourcename, database))
	}

	data.SetId(getAzureExternalDatasourceID(data))

	logger.Info().Msgf("updated external data source [%s] on database [%s]", datasourcename, database)

	return resourceAzureExternalDatasourceRead(ctx, data, meta)
}

func resourceAzureExternalDatasourceDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteAzureExternalDatasource(ctx, database, datasourcename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete external data source [%s] on database [%s]", datasourcename, database))
	}

	data.SetId("")

	logger.Info().Msgf("deleted external data source [%s] on database [%s]", datasourcename, database)

	return nil
}

func resourceAzureExternalDatasourceImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "azureexternaldatasource", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err := data.Set(serverProp, server); err != nil {
		return nil, err
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		return nil, errors.New("invalid ID")
	}
	if err = data.Set(databaseProp, parts[1]); err != nil {
		return nil, err
	}
	if err = data.Set(datasourcenameProp, parts[2]); err != nil {
		return nil, err
	}

	data.SetId(getAzureExternalDatasourceID(data))

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getAzureExternalDatasourceConnector(meta, data)
	if err != nil {
		return nil, err
	}

	extdatasource, err := connector.GetAzureExternalDatasource(ctx, database, datasourcename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to import external data source [%s] on database [%s]",datasourcename, database)
	}

	if extdatasource == nil {
		return nil, errors.Errorf("no external data source found [%s] on database [%s] for import",datasourcename, database)
	}

	if err = data.Set(datasourcenameProp, extdatasource.DataSourceName); err != nil {
		return nil, err
	}
	if err = data.Set(locationProp, extdatasource.Location); err != nil {
		return nil, err
	}
	if err = data.Set(typedescProp, extdatasource.TypeDesc); err != nil {
		return nil, err
	}
	if err = data.Set(credentialNameProp, extdatasource.CredentialName); err != nil {
		return nil, err
	}
	if err = data.Set(rdatabasenameProp, extdatasource.RDatabaseName); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getAzureExternalDatasourceConnector(meta interface{}, data *schema.ResourceData) (AzureExternalDatasourceConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(AzureExternalDatasourceConnector), nil
}

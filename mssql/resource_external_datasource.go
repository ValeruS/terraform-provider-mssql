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

func resourceExternalDatasource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalDatasourceCreate,
		ReadContext:   resourceExternalDatasourceRead,
		UpdateContext: resourceExternalDatasourceUpdate,
		DeleteContext: resourceExternalDatasourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceExternalDatasourceImport,
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
				Required: true,
			},
			credentialNameProp: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validate.SQLIdentifierName,
			},
			typedescProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validate.SQLExternalDatasourceType,
			},
			rdatabasenameProp: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validate.SQLIdentifierName,
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

type ExternalDatasourceConnector interface {
	CreateExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, typedesc, rdatabasename string) error
	GetExternalDatasource(ctx context.Context, database, credentialname string) (*model.ExternalDatasource, error)
	UpdateExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, rdatabasename string) error
	DeleteExternalDatasource(ctx context.Context, database, credentialname string) error
}

func resourceExternalDatasourceCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "externaldatasource", "create")
	logger.Debug().Msgf("Create %s", getExternalDatasourceID(data))

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)
	location := data.Get(locationProp).(string)
	credentialname := data.Get(credentialNameProp).(string)
	typedesc := data.Get(typedescProp).(string)
	rdatabasename := data.Get(rdatabasenameProp).(string)

	if (rdatabasename == "") && (typedesc == "RDBMS") {
		return diag.Errorf(rdatabasenameProp + " cannot be empty")
	}

	connector, err := getExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateExternalDatasource(ctx, database, datasourcename, location, credentialname, typedesc, rdatabasename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create external data source [%s] on database [%s]", datasourcename, database))
	}

	data.SetId(getExternalDatasourceID(data))

	logger.Info().Msgf("created external data source [%s] on database [%s]", datasourcename, database)

	return resourceExternalDatasourceRead(ctx, data, meta)
}

func resourceExternalDatasourceRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	}

	return nil
}

func resourceExternalDatasourceUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "externaldatasource", "update")
	logger.Debug().Msgf("Update %s", getDatabaseCredentialID(data))

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)
	location := data.Get(locationProp).(string)
	credentialname := data.Get(credentialNameProp).(string)
	rdatabasename := data.Get(rdatabasenameProp).(string)

	connector, err := getExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.UpdateExternalDatasource(ctx, database, datasourcename, location, credentialname, rdatabasename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to update external data source [%s] on database [%s]", datasourcename, database))
	}

	data.SetId(getExternalDatasourceID(data))

	logger.Info().Msgf("updated external data source [%s] on database [%s]", datasourcename, database)

	return resourceExternalDatasourceRead(ctx, data, meta)
}

func resourceExternalDatasourceDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "externaldatasource", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	connector, err := getExternalDatasourceConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteExternalDatasource(ctx, database, datasourcename); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete external data source [%s] on database [%s]", datasourcename, database))
	}

	logger.Info().Msgf("deleted external data source [%s] on database [%s]", datasourcename, database)

	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
	data.SetId("")

	return nil
}

func resourceExternalDatasourceImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "externaldatasource", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err = data.Set(serverProp, server); err != nil {
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

	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)

	data.SetId(getExternalDatasourceID(data))

	connector, err := getExternalDatasourceConnector(meta, data)
	if err != nil {
		return nil, err
	}

	datasource, err := connector.GetExternalDatasource(ctx, database, datasourcename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to import external data source [%s] on database [%s]", datasourcename, database)
	}

	if datasource == nil {
		return nil, errors.Errorf("no external data source found [%s] on database [%s] for import", datasourcename, database)
	}

	if err = data.Set(databaseProp, datasource.DatabaseName); err != nil {
		return nil, err
	}
	if err = data.Set(datasourcenameProp, datasource.DataSourceName); err != nil {
		return nil, err
	}
	if err = data.Set(datasourceIdProp, datasource.DataSourceId); err != nil {
		return nil, err
	}
	if err = data.Set(locationProp, datasource.Location); err != nil {
		return nil, err
	}
	if err = data.Set(typedescProp, datasource.TypeDesc); err != nil {
		return nil, err
	}
	if err = data.Set(credentialNameProp, datasource.CredentialName); err != nil {
		return nil, err
	}
	if err = data.Set(credentialIdProp, datasource.CredentialId); err != nil {
		return nil, err
	}
	if err = data.Set(rdatabasenameProp, datasource.RDatabaseName); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getExternalDatasourceConnector(meta interface{}, data *schema.ResourceData) (ExternalDatasourceConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(ExternalDatasourceConnector), nil
}

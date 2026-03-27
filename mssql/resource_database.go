package mssql

import (
	"context"
	"regexp"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		UpdateContext: resourceDatabaseUpdate,
		DeleteContext: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseImport,
		},
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			dbNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			collationProp: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
					"collation must only contain letters, numbers, and underscores",
				),
			},
			dbIDProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: defaultTimeout,
			Read:   defaultTimeout,
			Update: defaultTimeout,
			Delete: defaultTimeout,
		},
	}
}

type DatabaseConnector interface {
	CreateDatabase(ctx context.Context, databaseName string, collation string) error
	GetDatabase(ctx context.Context, databaseName string) (*model.Database, error)
	UpdateDatabaseCollation(ctx context.Context, databaseName string, collation string) error
	DeleteDatabase(ctx context.Context, databaseName string) error
}

func resourceDatabaseCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "create")
	logger.Debug().Msgf("Create %s", getDatabaseID(data))

	name := data.Get(dbNameProp).(string)
	collation := data.Get(collationProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateDatabase(ctx, name, collation); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create database [%s]", name))
	}

	data.SetId(getDatabaseID(data))

	logger.Info().Msgf("created database [%s]", name)

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	name := data.Get(dbNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	db, err := connector.GetDatabase(ctx, name)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get database [%s]", name))
	}

	if db == nil {
		logger.Info().Msgf("database [%s] does not exist", name)
		data.SetId("")
	} else {
		if err = data.Set(dbIDProp, db.DatabaseID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(dbNameProp, db.DatabaseName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(collationProp, db.Collation); err != nil {
			return diag.FromErr(err)
		}
	}

	logger.Info().Msgf("read database [%s]", name)

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	name := data.Get(dbNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.HasChange(collationProp) {
		oldCollation, _ := data.GetChange(collationProp)
		newCollation := data.Get(collationProp).(string)
		if err = connector.UpdateDatabaseCollation(ctx, name, newCollation); err != nil {
			if setErr := data.Set(collationProp, oldCollation); setErr != nil {
				logger.Error().Err(setErr).Msg("Failed to revert collation state after update error")
			}
			return diag.FromErr(errors.Wrapf(err, "unable to update collation for database [%s]", name))
		}
	}

	logger.Info().Msgf("updated database [%s]", name)

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	name := data.Get(dbNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteDatabase(ctx, name); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete database [%s]", name))
	}

	data.SetId("")

	logger.Info().Msgf("deleted database [%s]", name)

	return nil
}

func resourceDatabaseImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "database", "import")
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
		return nil, errors.New("invalid ID: expected sqlserver://host:port/database/db_name")
	}
	if err = data.Set(dbNameProp, parts[2]); err != nil {
		return nil, err
	}

	data.SetId(getDatabaseID(data))

	name := data.Get(dbNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return nil, err
	}

	db, err := connector.GetDatabase(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get database [%s]", name)
	}

	if db == nil {
		return nil, errors.Errorf("database [%s] does not exist", name)
	}

	if err = data.Set(dbIDProp, db.DatabaseID); err != nil {
		return nil, err
	}
	if err = data.Set(collationProp, db.Collation); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getDatabaseConnector(meta interface{}, data *schema.ResourceData) (DatabaseConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseConnector), nil
}

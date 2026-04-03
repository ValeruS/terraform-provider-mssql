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
			databaseNameProp: {
				Type:         schema.TypeString,
				Required:     true,
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
			databaseIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			compatibilityLevelProp: {
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
	GetMSSQLVersion(ctx context.Context) (string, error)
	CreateDatabase(ctx context.Context, databaseName string, collation string) error
	GetDatabase(ctx context.Context, databaseName string) (*model.Database, error)
	UpdateDatabase(ctx context.Context, databaseName string, newDatabaseName string, collation string) error
	DeleteDatabase(ctx context.Context, databaseName string) error
	DatabaseExists(ctx context.Context, databaseName string) (bool, error)
}

func resourceDatabaseCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "create")
	logger.Debug().Msgf("Create %s", getDatabaseID(data))

	databaseName := data.Get(databaseNameProp).(string)
	collationName := data.Get(collationProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	version, err := connector.GetMSSQLVersion(ctx)
	if err != nil {
		return diag.FromErr(errors.Wrap(err, "unable to get SQL Server version"))
	}
	if strings.Contains(version, "Microsoft SQL Azure") {
		return diag.Errorf("mssql_database is not supported on Azure SQL Database. " +
			"Use the azurerm_mssql_database resource from the AzureRM provider to manage Azure SQL Database databases. " +
			"This resource supports AWS RDS SQL Server, Azure SQL Managed Instance, and on-premises SQL Server.")
	}

	if err = connector.CreateDatabase(ctx, databaseName, collationName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create database [%s]", databaseName))
	}

	data.SetId(getDatabaseID(data))

	logger.Info().Msgf("created database [%s]", databaseName)

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	databaseName := data.Get(databaseNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if database exists
	exists, err := connector.DatabaseExists(ctx, databaseName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to check if database [%s] exists", databaseName))
	}
	if !exists {
		logger.Info().Msgf("Database [%s] does not exist", databaseName)
		data.SetId("")
		return nil
	}

	db, err := connector.GetDatabase(ctx, databaseName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get database [%s]", databaseName))
	}

	if db == nil {
		logger.Info().Msgf("database [%s] does not exist", databaseName)
		data.SetId("")
	} else {
		if err = data.Set(databaseIdProp, db.DatabaseID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(databaseNameProp, db.DatabaseName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(collationProp, db.Collation); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(compatibilityLevelProp, db.CompatibilityLevel); err != nil {
			return diag.FromErr(err)
		}
	}

	logger.Info().Msgf("read database [%s]", databaseName)

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	oldValue, newValue := data.GetChange(databaseNameProp)
	databaseName := oldValue.(string)
	newDatabaseName := newValue.(string)

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	for _, prop := range []string{databaseNameProp, collationProp} {
		if data.HasChange(prop) {
			oldValue, _ := data.GetChange(prop)
			oldValues[prop] = oldValue
		}
	}

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.HasChange(databaseNameProp) {
		oldDatabaseName := oldValues[databaseNameProp].(string)
		if err = connector.UpdateDatabase(ctx, oldDatabaseName, newDatabaseName, ""); err != nil {
			if setErr := data.Set(databaseNameProp, oldDatabaseName); setErr != nil {
				logger.Error().Err(setErr).Msg("Failed to revert databaseName state after update error")
			}
			return diag.FromErr(errors.Wrapf(err, "unable to update database [%s]", databaseName))
		}
		databaseName = newDatabaseName
	}

	if data.HasChange(collationProp) {
		collationName := data.Get(collationProp).(string)
		oldCollationName := oldValues[collationProp].(string)
		if err = connector.UpdateDatabase(ctx, databaseName, "", collationName); err != nil {
			if setErr := data.Set(collationProp, oldCollationName); setErr != nil {
				logger.Error().Err(setErr).Msg("Failed to revert collation state after update error")
			}
			return diag.FromErr(errors.Wrapf(err, "unable to update database [%s] collation", databaseName))
		}
	}

	data.SetId(getDatabaseID(data))

	logger.Info().Msgf("updated database [%s]", databaseName)

	return resourceDatabaseRead(ctx, data, meta)
}

func resourceDatabaseDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	databaseName := data.Get(databaseNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteDatabase(ctx, databaseName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete database [%s]", databaseName))
	}

	data.SetId("")

	logger.Info().Msgf("deleted database [%s]", databaseName)

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
	if err = data.Set(databaseNameProp, parts[2]); err != nil {
		return nil, err
	}

	data.SetId(getDatabaseID(data))

	databaseName := data.Get(databaseNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return nil, err
	}

	db, err := connector.GetDatabase(ctx, databaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get database [%s]", databaseName)
	}

	if db == nil {
		return nil, errors.Errorf("database [%s] does not exist", databaseName)
	}

	if err = data.Set(databaseIdProp, db.DatabaseID); err != nil {
		return nil, err
	}
	if err = data.Set(collationProp, db.Collation); err != nil {
		return nil, err
	}
	if err = data.Set(compatibilityLevelProp, db.CompatibilityLevel); err != nil {
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

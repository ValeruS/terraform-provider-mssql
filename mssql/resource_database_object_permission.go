package mssql

import (
	"context"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceDatabaseObjectPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseObjectPermissionCreate,
		ReadContext:   resourceDatabaseObjectPermissionRead,
		UpdateContext: resourceDatabaseObjectPermissionUpdate,
		DeleteContext: resourceDatabaseObjectPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseObjectPermissionImport,
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
			databaseProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			schemaNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			objectNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			columnNameProp: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			principalNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			principalIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			objectIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			permissionsProp: {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.SQLIdentifierPermission,
				},
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

type DatabaseObjectPermissionConnector interface {
	CreateDatabaseObjectPermission(ctx context.Context, permission *model.DatabaseObjectPermission) error
	GetDatabaseObjectPermission(ctx context.Context, database, schemaName, objectName, columnName, principalName string) (*model.DatabaseObjectPermission, error)
	UpdateDatabaseObjectPermission(ctx context.Context, database, schemaName, objectName, columnName, principalName string, permissions []string, changeType string) error
	DeleteDatabaseObjectPermission(ctx context.Context, permission *model.DatabaseObjectPermission) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceDatabaseObjectPermissionCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseobjectpermission", "create")
	logger.Debug().Msgf("Create %s", getDatabaseObjectPermissionID(data))

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	objectName := data.Get(objectNameProp).(string)
	columnName := data.Get(columnNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	permissions := toStringSlice(data.Get(permissionsProp).(*schema.Set).List())

	connector, err := getDatabaseObjectPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	permissionModel := &model.DatabaseObjectPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		ObjectName:    objectName,
		ColumnName:    columnName,
		PrincipalName: principalName,
		Permissions:   permissions,
	}
	if err = connector.CreateDatabaseObjectPermission(ctx, permissionModel); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to grant permissions [%s] on object [%s].[%s] in database [%s] to principal [%s]", strings.Join(permissions, ", "), schemaName, objectName, database, principalName))
	}

	data.SetId(getDatabaseObjectPermissionID(data))

	logger.Info().Msgf("granted permissions [%s] on object [%s].[%s] in database [%s] to principal [%s]", strings.Join(permissions, ", "), schemaName, objectName, database, principalName)

	return resourceDatabaseObjectPermissionRead(ctx, data, meta)
}

func resourceDatabaseObjectPermissionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseobjectpermission", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	objectName := data.Get(objectNameProp).(string)
	columnName := data.Get(columnNameProp).(string)
	principalName := data.Get(principalNameProp).(string)

	connector, err := getDatabaseObjectPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if database exists
	exists, err := connector.DatabaseExists(ctx, database)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to check if database [%s] exists", database))
	}
	if !exists {
		logger.Info().Msgf("Database [%s] does not exist", database)
		data.SetId("")
		return nil
	}

	permissions, err := connector.GetDatabaseObjectPermission(ctx, database, schemaName, objectName, columnName, principalName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read permissions on object [%s].[%s] in database [%s] for principal [%s]", schemaName, objectName, database, principalName))
	}
	if permissions == nil {
		logger.Info().Msgf("Principal [%s], object [%s].[%s] or column [%s] does not exist in database [%s]", principalName, schemaName, objectName, columnName, database)
		data.SetId("")
		return nil
	}

	if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(schemaNameProp, permissions.SchemaName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(objectNameProp, permissions.ObjectName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(columnNameProp, permissions.ColumnName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(principalNameProp, permissions.PrincipalName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(objectIdProp, permissions.ObjectID); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatabaseObjectPermissionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseobjectpermission", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	objectName := data.Get(objectNameProp).(string)
	columnName := data.Get(columnNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	oldVal, newVal := data.GetChange(permissionsProp)
	oldPermissions := oldVal.(*schema.Set)
	newPermissions := newVal.(*schema.Set)

	toGrant, toRevoke := stringSetDiff(oldPermissions, newPermissions)

	connector, err := getDatabaseObjectPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(toGrant) > 0 {
		if err = connector.UpdateDatabaseObjectPermission(ctx, database, schemaName, objectName, columnName, principalName, toGrant, "GRANT"); err != nil {
			if setErr := data.Set(permissionsProp, oldPermissions.List()); setErr != nil {
				logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", permissionsProp)
			}
			return diag.FromErr(errors.Wrapf(err, "unable to grant permissions [%s] on object [%s].[%s] in database [%s] to principal [%s]", strings.Join(toGrant, ", "), schemaName, objectName, database, principalName))
		}
	}
	if len(toRevoke) > 0 {
		if err = connector.UpdateDatabaseObjectPermission(ctx, database, schemaName, objectName, columnName, principalName, toRevoke, "REVOKE"); err != nil {
			if setErr := data.Set(permissionsProp, oldPermissions.List()); setErr != nil {
				logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", permissionsProp)
			}
			return diag.FromErr(errors.Wrapf(err, "unable to revoke permissions [%s] on object [%s].[%s] in database [%s] from principal [%s]", strings.Join(toRevoke, ", "), schemaName, objectName, database, principalName))
		}
	}

	logger.Info().Msgf("updated permissions on object [%s].[%s] in database [%s] for principal [%s]", schemaName, objectName, database, principalName)

	return resourceDatabaseObjectPermissionRead(ctx, data, meta)
}

func resourceDatabaseObjectPermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseobjectpermission", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	objectName := data.Get(objectNameProp).(string)
	columnName := data.Get(columnNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	permissions := toStringSlice(data.Get(permissionsProp).(*schema.Set).List())

	connector, err := getDatabaseObjectPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	permissionModel := &model.DatabaseObjectPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		ObjectName:    objectName,
		ColumnName:    columnName,
		PrincipalName: principalName,
		Permissions:   permissions,
	}
	if err = connector.DeleteDatabaseObjectPermission(ctx, permissionModel); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to revoke permissions [%s] on object [%s].[%s] in database [%s] from principal [%s]", strings.Join(permissions, ", "), schemaName, objectName, database, principalName))
	}

	data.SetId("")

	logger.Info().Msgf("revoked permissions [%s] on object [%s].[%s] in database [%s] from principal [%s]", strings.Join(permissions, ", "), schemaName, objectName, database, principalName)

	return nil
}

func resourceDatabaseObjectPermissionImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "databaseobjectpermission", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err = data.Set(serverProp, server); err != nil {
		return nil, err
	}

	// Path: /<database>/object_permission/<schema_name>/<object_name>/<principal_name>[/<column_name>]
	parts := strings.Split(u.Path, "/")
	if (len(parts) != 6 && len(parts) != 7) || parts[2] != "object_permission" {
		return nil, errors.New("invalid ID, expected <scheme>://<host>[:<port>]/<database>/object_permission/<schema_name>/<object_name>/<principal_name>[/<column_name>]")
	}

	if err = data.Set(databaseProp, parts[1]); err != nil {
		return nil, err
	}
	if err = data.Set(schemaNameProp, parts[3]); err != nil {
		return nil, err
	}
	if err = data.Set(objectNameProp, parts[4]); err != nil {
		return nil, err
	}
	if err = data.Set(principalNameProp, parts[5]); err != nil {
		return nil, err
	}
	if len(parts) == 7 {
		if err = data.Set(columnNameProp, parts[6]); err != nil {
			return nil, err
		}
	}

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	objectName := data.Get(objectNameProp).(string)
	columnName := data.Get(columnNameProp).(string)
	principalName := data.Get(principalNameProp).(string)

	data.SetId(getDatabaseObjectPermissionID(data))

	connector, err := getDatabaseObjectPermissionConnector(meta, data)
	if err != nil {
		return nil, err
	}

	permissions, err := connector.GetDatabaseObjectPermission(ctx, database, schemaName, objectName, columnName, principalName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to import permissions on object [%s].[%s] in database [%s] for principal [%s]", schemaName, objectName, database, principalName)
	}
	if permissions == nil {
		return nil, errors.Errorf("principal [%s], object [%s].[%s] or column [%s] does not exist in database [%s]", principalName, schemaName, objectName, columnName, database)
	}
	if len(permissions.Permissions) == 0 {
		return nil, errors.Errorf("no permissions found on object [%s].[%s] in database [%s] for principal [%s]", schemaName, objectName, database, principalName)
	}

	if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
		return nil, err
	}
	if err = data.Set(objectIdProp, permissions.ObjectID); err != nil {
		return nil, err
	}
	if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getDatabaseObjectPermissionConnector(meta interface{}, data *schema.ResourceData) (DatabaseObjectPermissionConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseObjectPermissionConnector), nil
}

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

func resourceDatabaseSchemaPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseSchemaPermissionCreate,
		ReadContext:   resourceDatabaseSchemaPermissionRead,
		UpdateContext: resourceDatabaseSchemaPermissionUpdate,
		DeleteContext: resourceDatabaseSchemaPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseSchemaPermissionImport,
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
			schemaIdProp: {
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

type DatabaseSchemaPermissionConnector interface {
	CreateDatabaseSchemaPermission(ctx context.Context, permission *model.DatabaseSchemaPermission) error
	GetDatabaseSchemaPermission(ctx context.Context, database, schemaName, principalName string) (*model.DatabaseSchemaPermission, error)
	UpdateDatabaseSchemaPermission(ctx context.Context, database, schemaName, principalName string, permissions []string, changeType string) error
	DeleteDatabaseSchemaPermission(ctx context.Context, permission *model.DatabaseSchemaPermission) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceDatabaseSchemaPermissionCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseschemapermission", "create")
	logger.Debug().Msgf("Create %s", getDatabaseSchemaPermissionID(data))

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	permissions := toStringSlice(data.Get(permissionsProp).(*schema.Set).List())

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	permissionModel := &model.DatabaseSchemaPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		PrincipalName: principalName,
		Permissions:   permissions,
	}
	if err = connector.CreateDatabaseSchemaPermission(ctx, permissionModel); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to grant permissions [%s] on schema [%s] in database [%s] to principal [%s]", strings.Join(permissions, ", "), schemaName, database, principalName))
	}

	data.SetId(getDatabaseSchemaPermissionID(data))

	logger.Info().Msgf("granted permissions [%s] on schema [%s] in database [%s] to principal [%s]", strings.Join(permissions, ", "), schemaName, database, principalName)

	return resourceDatabaseSchemaPermissionRead(ctx, data, meta)
}

func resourceDatabaseSchemaPermissionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseschemapermission", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
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

	permissions, err := connector.GetDatabaseSchemaPermission(ctx, database, schemaName, principalName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read permissions on schema [%s] in database [%s] for principal [%s]", schemaName, database, principalName))
	}
	if permissions == nil {
		logger.Info().Msgf("Principal [%s] or schema [%s] does not exist in database [%s]", principalName, schemaName, database)
		data.SetId("")
		return nil
	}

	if err = data.Set(databaseProp, permissions.DatabaseName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(schemaNameProp, permissions.SchemaName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(principalNameProp, permissions.PrincipalName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(schemaIdProp, permissions.SchemaID); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatabaseSchemaPermissionUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseschemapermission", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	oldVal, newVal := data.GetChange(permissionsProp)
	oldPermissions := oldVal.(*schema.Set)
	newPermissions := newVal.(*schema.Set)

	toGrant, toRevoke := stringSetDiff(oldPermissions, newPermissions)

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(toGrant) > 0 {
		if err = connector.UpdateDatabaseSchemaPermission(ctx, database, schemaName, principalName, toGrant, "GRANT"); err != nil {
			if setErr := data.Set(permissionsProp, oldPermissions.List()); setErr != nil {
				logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", permissionsProp)
			}
			return diag.FromErr(errors.Wrapf(err, "unable to grant permissions [%s] on schema [%s] in database [%s] to principal [%s]", strings.Join(toGrant, ", "), schemaName, database, principalName))
		}
	}
	if len(toRevoke) > 0 {
		if err = connector.UpdateDatabaseSchemaPermission(ctx, database, schemaName, principalName, toRevoke, "REVOKE"); err != nil {
			if setErr := data.Set(permissionsProp, oldPermissions.List()); setErr != nil {
				logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", permissionsProp)
			}
			return diag.FromErr(errors.Wrapf(err, "unable to revoke permissions [%s] on schema [%s] in database [%s] from principal [%s]", strings.Join(toRevoke, ", "), schemaName, database, principalName))
		}
	}

	logger.Info().Msgf("updated permissions on schema [%s] in database [%s] for principal [%s]", schemaName, database, principalName)

	return resourceDatabaseSchemaPermissionRead(ctx, data, meta)
}

func resourceDatabaseSchemaPermissionDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseschemapermission", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)
	permissions := toStringSlice(data.Get(permissionsProp).(*schema.Set).List())

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	permissionModel := &model.DatabaseSchemaPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		PrincipalName: principalName,
		Permissions:   permissions,
	}
	if err = connector.DeleteDatabaseSchemaPermission(ctx, permissionModel); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to revoke permissions [%s] on schema [%s] in database [%s] from principal [%s]", strings.Join(permissions, ", "), schemaName, database, principalName))
	}

	data.SetId("")

	logger.Info().Msgf("revoked permissions [%s] on schema [%s] in database [%s] from principal [%s]", strings.Join(permissions, ", "), schemaName, database, principalName)

	return nil
}

func resourceDatabaseSchemaPermissionImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "databaseschemapermission", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err = data.Set(serverProp, server); err != nil {
		return nil, err
	}

	// Path: /<database>/schema_permission/<schema_name>/<principal_name>
	parts := strings.Split(u.Path, "/")
	if len(parts) != 5 || parts[2] != "schema_permission" {
		return nil, errors.New("invalid ID, expected <scheme>://<host>[:<port>]/<database>/schema_permission/<schema_name>/<principal_name>")
	}

	if err = data.Set(databaseProp, parts[1]); err != nil {
		return nil, err
	}
	if err = data.Set(schemaNameProp, parts[3]); err != nil {
		return nil, err
	}
	if err = data.Set(principalNameProp, parts[4]); err != nil {
		return nil, err
	}

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)

	data.SetId(getDatabaseSchemaPermissionID(data))

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
	if err != nil {
		return nil, err
	}

	permissions, err := connector.GetDatabaseSchemaPermission(ctx, database, schemaName, principalName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to import permissions on schema [%s] in database [%s] for principal [%s]", schemaName, database, principalName)
	}
	if permissions == nil {
		return nil, errors.Errorf("principal [%s] or schema [%s] does not exist in database [%s]", principalName, schemaName, database)
	}
	if len(permissions.Permissions) == 0 {
		return nil, errors.Errorf("no permissions found on schema [%s] in database [%s] for principal [%s]", schemaName, database, principalName)
	}

	if err = data.Set(principalIdProp, permissions.PrincipalID); err != nil {
		return nil, err
	}
	if err = data.Set(schemaIdProp, permissions.SchemaID); err != nil {
		return nil, err
	}
	if err = data.Set(permissionsProp, permissions.Permissions); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getDatabaseSchemaPermissionConnector(meta interface{}, data *schema.ResourceData) (DatabaseSchemaPermissionConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseSchemaPermissionConnector), nil
}

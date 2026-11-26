package mssql

import (
	"context"

	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func dataSourceDatabaseObjectPermission() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseObjectPermissionRead,
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
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: defaultTimeout,
		},
	}
}

func dataSourceDatabaseObjectPermissionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	permissions, err := connector.GetDatabaseObjectPermission(ctx, database, schemaName, objectName, columnName, principalName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read permissions on object [%s].[%s] in database [%s] for principal [%s]", schemaName, objectName, database, principalName))
	}
	if permissions == nil {
		return diag.Errorf("principal [%s], object [%s].[%s] or column [%s] does not exist in database [%s]", principalName, schemaName, objectName, columnName, database)
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
	data.SetId(getDatabaseObjectPermissionID(data))

	return nil
}

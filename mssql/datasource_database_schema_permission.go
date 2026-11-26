package mssql

import (
	"context"

	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func dataSourceDatabaseSchemaPermission() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseSchemaPermissionRead,
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

func dataSourceDatabaseSchemaPermissionRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databaseschemapermission", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	principalName := data.Get(principalNameProp).(string)

	connector, err := getDatabaseSchemaPermissionConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	permissions, err := connector.GetDatabaseSchemaPermission(ctx, database, schemaName, principalName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read permissions on schema [%s] in database [%s] for principal [%s]", schemaName, database, principalName))
	}
	if permissions == nil {
		return diag.Errorf("principal [%s] or schema [%s] does not exist in database [%s]", principalName, schemaName, database)
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
	data.SetId(getDatabaseSchemaPermissionID(data))

	return nil
}

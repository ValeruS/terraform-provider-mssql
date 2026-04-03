package mssql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func dataSourceDatabase() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseRead,
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
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
			},
			collationProp: {
				Type:     schema.TypeString,
				Computed: true,
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
			Read: defaultTimeout,
		},
	}
}

func dataSourceDatabaseRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "database", "read")
	logger.Debug().Msgf("Read datasource %s", data.Id())

	databaseName := data.Get(databaseNameProp).(string)

	connector, err := getDatabaseConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	db, err := connector.GetDatabase(ctx, databaseName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get database [%s]", databaseName))
	}

	if db == nil {
		return diag.Errorf("database [%s] does not exist", databaseName)
	}

	if err = data.Set(databaseNameProp, db.DatabaseName); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(databaseIdProp, db.DatabaseID); err != nil {
		return diag.FromErr(err)
	}

	if err = data.Set(collationProp, db.Collation); err != nil {
		return diag.FromErr(err)
	}
	if err = data.Set(compatibilityLevelProp, db.CompatibilityLevel); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(getDatabaseID(data))

	return nil
}

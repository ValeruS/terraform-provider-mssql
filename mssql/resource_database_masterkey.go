package mssql

import (
	"context"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceDatabaseMasterkey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseMasterkeyCreate,
		ReadContext:   resourceDatabaseMasterkeyRead,
		UpdateContext: resourceDatabaseMasterkeyUpdate,
		DeleteContext: resourceDatabaseMasterkeyDelete,
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
			passwordProp: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validate.SQLIdentifierPassword,
				ExactlyOneOf: []string{passwordProp, passwordWoProp},
			},
			passwordWoProp: {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				WriteOnly:     true,
				ValidateFunc:  validate.SQLIdentifierPassword,
				ConflictsWith: []string{passwordProp},
				RequiredWith:  []string{passwordWoVersionProp},
			},
			passwordWoVersionProp: {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{passwordWoProp},
			},
			keynameProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyguidProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			principalIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			symmetrickeyidProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			keylengthProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			keyalgorithmProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			algorithmdescProp: {
				Type:     schema.TypeString,
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

type DatabaseMasterkeyConnector interface {
	CreateDatabaseMasterkey(ctx context.Context, database, password string) error
	GetDatabaseMasterkey(ctx context.Context, database string) (*model.DatabaseMasterkey, error)
	UpdateDatabaseMasterkey(ctx context.Context, database, password string) error
	DeleteDatabaseMasterkey(ctx context.Context, database string) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

func resourceDatabaseMasterkeyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "create")
	logger.Debug().Msgf("Create %s", getDatabaseMasterkeyID(data))

	database := data.Get(databaseProp).(string)
	password := data.Get(passwordProp).(string)
	if password == "" {
		// password_wo is never stored in state, so it is read from the raw configuration
		wo, diags := data.GetRawConfigAt(cty.GetAttrPath(passwordWoProp))
		if diags.HasError() {
			return diags
		}
		if wo.IsKnown() && !wo.IsNull() {
			password = wo.AsString()
		}
	}

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateDatabaseMasterkey(ctx, database, password); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create database master key on database [%s]", database))
	}

	data.SetId(getDatabaseMasterkeyID(data))

	logger.Info().Msgf("created database master key on database [%s]", database)

	return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
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

	masterkey, err := connector.GetDatabaseMasterkey(ctx, database)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read database master key on database [%s]", database))
	}

	if masterkey == nil {
		logger.Info().Msgf("No database master key found on database [%s]", database)
		data.SetId("")
	} else {
		if err = data.Set(keynameProp, masterkey.KeyName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(principalIdProp, masterkey.PrincipalID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(symmetrickeyidProp, masterkey.SymmetricKeyID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keylengthProp, masterkey.KeyLength); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keyalgorithmProp, masterkey.KeyAlgorithm); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(algorithmdescProp, masterkey.AlgorithmDesc); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(keyguidProp, masterkey.KeyGuid); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatabaseMasterkeyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	// Update also runs when only the server block changes, for example when the
	// credentials used to connect are rotated. The master key is only regenerated
	// when password or password_wo_version changes.
	if !data.HasChange(passwordProp) && !data.HasChange(passwordWoVersionProp) {
		return resourceDatabaseMasterkeyRead(ctx, data, meta)
	}

	database := data.Get(databaseProp).(string)
	password := data.Get(passwordProp).(string)
	if password == "" {
		// password_wo is never stored in state, so it is read from the raw configuration
		wo, diags := data.GetRawConfigAt(cty.GetAttrPath(passwordWoProp))
		if diags.HasError() {
			return diags
		}
		if wo.IsKnown() && !wo.IsNull() {
			password = wo.AsString()
		}
	}

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	for _, prop := range []string{passwordProp, passwordWoVersionProp} {
		if data.HasChange(prop) {
			oldValue, _ := data.GetChange(prop)
			oldValues[prop] = oldValue
		}
	}

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.UpdateDatabaseMasterkey(ctx, database, password); err != nil {
		// If update fails, revert all changed values in the state
		for prop, oldValue := range oldValues {
			if err := data.Set(prop, oldValue); err != nil {
				logger.Error().Err(err).Msgf("Failed to revert %s state after update error", prop)
			}
		}
		return diag.FromErr(errors.Wrapf(err, "unable to update database key on database [%s]", database))
	}

	data.SetId(getDatabaseMasterkeyID(data))

	logger.Info().Msgf("updated database master key on database [%s]", database)

	return resourceDatabaseMasterkeyRead(ctx, data, meta)
}

func resourceDatabaseMasterkeyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "databasemasterkey", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)

	connector, err := getDatabaseMasterkeyConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteDatabaseMasterkey(ctx, database); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete database master key on database [%s]", database))
	}

	data.SetId("")

	logger.Info().Msgf("deleted database master key on database [%s]", database)

	return nil
}

func getDatabaseMasterkeyConnector(meta interface{}, data *schema.ResourceData) (DatabaseMasterkeyConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseMasterkeyConnector), nil
}

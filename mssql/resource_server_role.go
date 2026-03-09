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

func resourceServerRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerRoleCreate,
		ReadContext:   resourceServerRoleRead,
		UpdateContext: resourceServerRoleUpdate,
		DeleteContext: resourceServerRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServerRoleImport,
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
			roleNameProp: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			ownerNameProp: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultSAPropDefault,
				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
					return (old == "" && new == defaultSAPropDefault) || (old == defaultSAPropDefault && new == "")
				},
			},
			ownerIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			principalIdProp: {
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

type ServerRoleConnector interface {
	CreateServerRole(ctx context.Context, roleName string, ownerName string) error
	GetServerRole(ctx context.Context, roleName string) (*model.ServerRole, error)
	UpdateServerRoleName(ctx context.Context, newroleName string, oldroleName string) error
	UpdateServerRoleOwner(ctx context.Context, roleName string, ownerName string) error
	DeleteServerRole(ctx context.Context, roleName string) error
}

func resourceServerRoleCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role", "create")
	logger.Debug().Msgf("Create %s", getServerRoleID(data))

	roleName := data.Get(roleNameProp).(string)
	ownerName := data.Get(ownerNameProp).(string)

	connector, err := getServerRoleConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateServerRole(ctx, roleName, ownerName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create role [%s]", roleName))
	}

	data.SetId(getServerRoleID(data))

	logger.Info().Msgf("created role [%s]", roleName)

	return resourceServerRoleRead(ctx, data, meta)
}

func resourceServerRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	roleName := data.Get(roleNameProp).(string)

	connector, err := getServerRoleConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := connector.GetServerRole(ctx, roleName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get role [%s]", roleName))
	}

	if role == nil {
		logger.Info().Msgf("role [%s] does not exist", roleName)
		data.SetId("")
	} else {
		if err = data.Set(principalIdProp, role.RoleID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(roleNameProp, role.RoleName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(ownerNameProp, role.OwnerName); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(ownerIdProp, role.OwnerId); err != nil {
			return diag.FromErr(err)
		}
	}

	logger.Info().Msgf("read role [%s]", roleName)

	return nil
}

func resourceServerRoleDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	roleName := data.Get(roleNameProp).(string)

	connector, err := getServerRoleConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteServerRole(ctx, roleName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete role [%s]", roleName))
	}

	data.SetId("")

	logger.Info().Msgf("deleted role [%s]", roleName)

	return nil
}

func resourceServerRoleUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	roleName := data.Get(roleNameProp).(string)
	ownerName := data.Get(ownerNameProp).(string)

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	for _, prop := range []string{roleNameProp, ownerNameProp} {
		if data.HasChange(prop) {
			oldValue, _ := data.GetChange(prop)
			oldValues[prop] = oldValue
		}
	}

	connector, err := getServerRoleConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if data.HasChange(roleNameProp) {
		oldRoleName := oldValues[roleNameProp].(string)
		if err = connector.UpdateServerRoleName(ctx, roleName, oldRoleName); err != nil {
			if setErr := data.Set(roleNameProp, oldRoleName); setErr != nil {
				logger.Error().Err(setErr).Msg("Failed to revert roleName state after update error")
			}
			return diag.FromErr(errors.Wrapf(err, "unable to update role name [%s]", roleName))
		}
	}
	if data.HasChange(ownerNameProp) {
		oldOwnerName := oldValues[ownerNameProp].(string)
		if err = connector.UpdateServerRoleOwner(ctx, roleName, ownerName); err != nil {
			if setErr := data.Set(ownerNameProp, oldOwnerName); setErr != nil {
				logger.Error().Err(setErr).Msg("Failed to revert ownerName state after update error")
			}
			return diag.FromErr(errors.Wrapf(err, "unable to update role owner [%s]", roleName))
		}
	}

	data.SetId(getServerRoleID(data))

	logger.Info().Msgf("updated role [%s]", roleName)

	return resourceServerRoleRead(ctx, data, meta)
}

func resourceServerRoleImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "role", "import")
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
	if err = data.Set(roleNameProp, parts[2]); err != nil {
		return nil, err
	}

	data.SetId(getServerRoleID(data))

	roleName := data.Get(roleNameProp).(string)

	connector, err := getServerRoleConnector(meta, data)
	if err != nil {
		return nil, err
	}

	role, err := connector.GetServerRole(ctx, roleName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get role [%s]", roleName)
	}

	if role == nil {
		return nil, errors.Errorf("role [%s] does not exist", roleName)
	}

	if err = data.Set(principalIdProp, role.RoleID); err != nil {
		return nil, err
	}
	if err = data.Set(ownerNameProp, role.OwnerName); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getServerRoleConnector(meta interface{}, data *schema.ResourceData) (ServerRoleConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(ServerRoleConnector), nil
}

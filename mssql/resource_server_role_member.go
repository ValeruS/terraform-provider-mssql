package mssql

import (
	"context"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func resourceServerRoleMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerRoleMemberCreate,
		ReadContext:   resourceServerRoleMemberRead,
		UpdateContext: resourceServerRoleMemberUpdate,
		DeleteContext: resourceServerRoleMemberDelete,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			membersProp: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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

type ServerRoleMemberConnector interface {
	CreateServerRoleMember(ctx context.Context, roleName string, members []string) error
	GetServerRoleMember(ctx context.Context, roleName string, managedMembers []string) (*model.ServerRoleMember, error)
	UpdateServerRoleMember(ctx context.Context, roleName string, members []string, changeType string) error
	DeleteServerRoleMember(ctx context.Context, roleName string, members []string) error
}

func resourceServerRoleMemberCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role_member", "create")
	logger.Debug().Msgf("Create %s", getServerRoleMemberID(data))

	roleName := data.Get(roleNameProp).(string)
	members := data.Get(membersProp).(*schema.Set).List()

	connector, err := getServerRoleMemberConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateServerRoleMember(ctx, roleName, toStringSlice(members)); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to add members [%s] to role [%s]", strings.Join(toStringSlice(members), ", "), roleName))
	}

	data.SetId(getServerRoleMemberID(data))

	logger.Info().Msgf("added members [%s] to role [%s]", strings.Join(toStringSlice(members), ", "), roleName)

	return resourceServerRoleMemberRead(ctx, data, meta)
}

func resourceServerRoleMemberRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role_member", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	roleName := data.Get(roleNameProp).(string)
	managedMembers := toStringSlice(data.Get(membersProp).(*schema.Set).List())

	connector, err := getServerRoleMemberConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	roleMembers, err := connector.GetServerRoleMember(ctx, roleName, managedMembers)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to get role members for role [%s]", roleName))
	}

	if roleMembers == nil {
		logger.Info().Msgf("role members for role [%s] do not exist", roleName)
		data.SetId("")
	} else {
		if err = data.Set(membersProp, roleMembers.Members); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceServerRoleMemberUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role_member", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	roleName := data.Get(roleNameProp).(string)
	oldVal, newVal := data.GetChange(membersProp)
	oldMembers := oldVal.(*schema.Set)
	newMembers := newVal.(*schema.Set)

	// Store old values for all properties that might change
	oldValues := make(map[string]interface{})
	if data.HasChange(membersProp) {
		oldValue, _ := data.GetChange(membersProp)
		if oldSet, ok := oldValue.(*schema.Set); ok {
			oldValues[membersProp] = oldSet.List()
		}
	}

	toAdd, toRemove := stringSetDiff(oldMembers, newMembers)

	connector, err := getServerRoleMemberConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(toAdd) > 0 {
		if err = connector.UpdateServerRoleMember(ctx, roleName, toAdd, "ADD"); err != nil {
			// If update fails, revert all changed values in the state
			for prop, oldValue := range oldValues {
				if setErr := data.Set(prop, oldValue); setErr != nil {
					logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", prop)
				}
			}
			return diag.FromErr(errors.Wrapf(err, "unable to add members to role [%s]", roleName))
		}
		logger.Info().Msgf("added members to role [%s]", roleName)
	}
	if len(toRemove) > 0 {
		if err = connector.UpdateServerRoleMember(ctx, roleName, toRemove, "DROP"); err != nil {
			// If update fails, revert all changed values in the state
			for prop, oldValue := range oldValues {
				if setErr := data.Set(prop, oldValue); setErr != nil {
					logger.Error().Err(setErr).Msgf("Failed to revert %s state after update error", prop)
				}
			}
			return diag.FromErr(errors.Wrapf(err, "unable to remove members from role [%s]", roleName))
		}
		logger.Info().Msgf("removed members from role [%s]", roleName)
	}

	data.SetId(getServerRoleMemberID(data))

	logger.Info().Msgf("updated role members for role [%s]", roleName)

	return resourceServerRoleMemberRead(ctx, data, meta)
}

func resourceServerRoleMemberDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "role_member", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	roleName := data.Get(roleNameProp).(string)
	managedMembers := toStringSlice(data.Get(membersProp).(*schema.Set).List())

	connector, err := getServerRoleMemberConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteServerRoleMember(ctx, roleName, managedMembers); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete role members for role [%s]", roleName))
	}

	data.SetId("")

	logger.Info().Msgf("deleted role members for role [%s]", roleName)

	return nil
}

func getServerRoleMemberConnector(meta interface{}, data *schema.ResourceData) (ServerRoleMemberConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(ServerRoleMemberConnector), nil
}

package mssql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func dataSourceServerRoleMember() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerRoleMemberRead,
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
			roleNameProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			membersProp: {
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

func dataSourceServerRoleMemberRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "server_role_member", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	roleName := data.Get(roleNameProp).(string)

	connector, err := getServerRoleMemberConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	members, err := connector.GetServerRoleMember(ctx, roleName, nil)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read server role members for role [%s]", roleName))
	}
	if members == nil {
		return diag.Errorf("No server role members found for role [%s]", roleName)
	} else {
		if err = data.Set(membersProp, members.Members); err != nil {
			return diag.FromErr(err)
		}
		data.SetId(getServerRoleMemberID(data))
	}

	return nil
}

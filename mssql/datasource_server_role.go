package mssql

import (
	"context"

	"github.com/ValeruS/terraform-provider-mssql/mssql/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func dataSourceServerRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerRoleRead,
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
				ForceNew:     true,
				ValidateFunc: validate.SQLIdentifier,
			},
			ownerNameProp: {
				Type:     schema.TypeString,
				Computed: true,
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
			Read: defaultTimeout,
		},
	}
}

func dataSourceServerRoleRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return diag.Errorf("The role [%s] does not exist", roleName)
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
		data.SetId(getServerRoleID(data))
	}

	return nil
}

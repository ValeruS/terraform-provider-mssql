package mssql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pkg/errors"
)

func dataSourceEntraIDLogin() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEntraIDLoginRead,
		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			loginNameProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			sidStrProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			principalIdProp: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			defaultDatabaseProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
			defaultLanguageProp: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Read: defaultTimeout,
		},
	}
}

func dataSourceEntraIDLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "EntraIDLogin", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	loginName := data.Get(loginNameProp).(string)

	connector, err := getEntraIDLoginConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	EntraIDLogin, err := connector.GetEntraIDLogin(ctx, loginName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to EntraID login [%s]", loginName))
	}
	if EntraIDLogin == nil {
		logger.Info().Msgf("No EntraID Login found for [%s]", loginName)
		data.SetId("")
	} else {
		if err = data.Set(principalIdProp, EntraIDLogin.PrincipalID); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(sidStrProp, EntraIDLogin.Sid); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultDatabaseProp, EntraIDLogin.DefaultDatabase); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultLanguageProp, EntraIDLogin.DefaultLanguage); err != nil {
			return diag.FromErr(err)
		}
		data.SetId(getLoginID(data))
	}

	return nil
}

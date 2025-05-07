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

func resourceEntraIDLogin() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEntraIDLoginCreate,
		ReadContext:   resourceEntraIDLoginRead,
		DeleteContext: resourceEntraIDLoginDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEntraIDLoginImport,
		},
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
			loginNameProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validate.SQLIdentifier,
			},
			objectIdProp: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			sidStrProp: {
				Type:     schema.TypeString,
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

type EntraIDLoginConnector interface {
	CreateEntraIDLogin(ctx context.Context, name, objectId string) error
	GetEntraIDLogin(ctx context.Context, name string) (*model.EntraIDLogin, error)
	DeleteEntraIDLogin(ctx context.Context, name string) error
}

func resourceEntraIDLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "EntraIDLogin", "create")
	logger.Debug().Msgf("Create %s", getLoginID(data))

	loginName := data.Get(loginNameProp).(string)
	objectId := data.Get(objectIdProp).(string)

	connector, err := getEntraIDLoginConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.CreateEntraIDLogin(ctx, loginName, objectId); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to create EntraID Login [%s]", loginName))
	}

	data.SetId(getLoginID(data))

	logger.Info().Msgf("created EntraID Login [%s]", loginName)

	return resourceEntraIDLoginRead(ctx, data, meta)
}

func resourceEntraIDLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "EntraIDLogin", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	loginName := data.Get(loginNameProp).(string)

	connector, err := getEntraIDLoginConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	EntraIDLogin, err := connector.GetEntraIDLogin(ctx, loginName)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to read EntraID Login [%s]", loginName))
	}
	if EntraIDLogin == nil {
		logger.Info().Msgf("No EntraID Login found for [%s]", loginName)
		data.SetId("")
	} else {
		if err = data.Set(defaultDatabaseProp, EntraIDLogin.DefaultDatabase); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(defaultLanguageProp, EntraIDLogin.DefaultLanguage); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(sidStrProp, EntraIDLogin.Sid); err != nil {
			return diag.FromErr(err)
		}
		if err = data.Set(principalIdProp, EntraIDLogin.PrincipalID); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceEntraIDLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "EntraIDLogin", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	loginName := data.Get(loginNameProp).(string)

	connector, err := getEntraIDLoginConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = connector.DeleteEntraIDLogin(ctx, loginName); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to delete EntraID Login [%s]", loginName))
	}

	logger.Info().Msgf("deleted EntraID Login [%s]", loginName)

	data.SetId("")

	return nil
}

func resourceEntraIDLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	logger := loggerFromMeta(meta, "EntraIDLogin", "import")
	logger.Debug().Msgf("Import %s", data.Id())

	server, u, err := serverFromId(data.Id())
	if err != nil {
		return nil, err
	}
	if err = data.Set(serverProp, server); err != nil {
		return nil, err
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		return nil, errors.New("invalid ID")
	}
	if err = data.Set(loginNameProp, parts[2]); err != nil {
		return nil, err
	}

	data.SetId(getLoginID(data))

	loginName := data.Get(loginNameProp).(string)

	connector, err := getEntraIDLoginConnector(meta, data)
	if err != nil {
		return nil, err
	}

	EntraIDLogin, err := connector.GetEntraIDLogin(ctx, loginName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read EntraID Login [%s] for import", loginName)
	}

	if EntraIDLogin == nil {
		return nil, errors.Errorf("no EntraID Login [%s] found for import", loginName)
	}

	if err = data.Set(objectIdProp, EntraIDLogin.ObjectId); err != nil {
		return nil, err
	}
	if err = data.Set(defaultDatabaseProp, EntraIDLogin.DefaultDatabase); err != nil {
		return nil, err
	}
	if err = data.Set(defaultLanguageProp, EntraIDLogin.DefaultLanguage); err != nil {
		return nil, err
	}
	if err = data.Set(sidStrProp, EntraIDLogin.Sid); err != nil {
		return nil, err
	}
	if err = data.Set(principalIdProp, EntraIDLogin.PrincipalID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func getEntraIDLoginConnector(meta interface{}, data *schema.ResourceData) (EntraIDLoginConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(EntraIDLoginConnector), nil
}

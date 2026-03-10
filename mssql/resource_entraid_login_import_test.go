package mssql

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccEntraIDLogin_Azure_BasicImport(t *testing.T) {
	clientUser := os.Getenv("TF_ACC_AZURE_USER_CLIENT_USER")
	loginType := "azure"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckEntraIDLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "test_import", loginType, map[string]interface{}{"login_name": clientUser}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExists("mssql_entraid_login.test_import"),
				),
			},
			{
				ResourceName:      "mssql_entraid_login.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_entraid_login.test_import", true),
			},
		},
	})
}

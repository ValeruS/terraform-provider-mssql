package mssql

import (
	"os"
	"regexp"
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
				ImportStateIdFunc: testAccImportStateId("mssql_entraid_login.test_import", "", "azure=true"),
			},
		},
	})
}

func TestAccEntraIDLogin_Azure_ObjectIdImport(t *testing.T) {
	clientUser := os.Getenv("TF_ACC_AZURE_USER_CLIENT_USER")
	objectId := os.Getenv("TF_ACC_AZURE_USER_CLIENT_OBJECTID")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckEntraIDLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "test_import", "azure", map[string]interface{}{"login_name": clientUser, "object_id": objectId}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExists("mssql_entraid_login.test_import"),
				),
			},
			{
				ResourceName:      "mssql_entraid_login.test_import",
				ImportState:       true,
				ImportStateIdFunc: testAccImportStateId("mssql_entraid_login.test_import", "not-a-guid", "azure=true"),
				ExpectError:       regexp.MustCompile(`expected "object_id" to be a valid UUID`),
			},
			{
				ResourceName:      "mssql_entraid_login.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_entraid_login.test_import", objectId, "azure=true"),
			},
		},
	})
}

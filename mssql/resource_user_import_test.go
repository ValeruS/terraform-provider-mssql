package mssql

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUser_Local_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckUser(t, "test_import", "login", map[string]interface{}{"username": "user_import", "login_name": "user_import", "login_password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("mssql_user.test_import"),
				),
			},
			{
				ResourceName:      "mssql_user.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_user.test_import", "", "azure=false"),
			},
		},
	})
}

func TestAccUser_AzureadChainEnvironment_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckUser(t, "test_import", "fedauth", map[string]interface{}{"database": "testdb", "username": "user_import_chain", "password": "valueIsH8kd$¡", "roles": "[\"db_owner\"]", "credential_type": "environment"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("mssql_user.test_import"),
				),
			},
			{
				// ImportStateVerify also checks that the imported server block matches
				// the azuread_default_chain_auth block of the configuration.
				ResourceName:            "mssql_user.test_import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
				ImportStateIdFunc:       testAccImportStateId("mssql_user.test_import", "", "auth=default_chain&credential_type=environment"),
			},
		},
	})
}

func TestAccUser_Azure_ObjectIdImport(t *testing.T) {
	clientId := os.Getenv("TF_ACC_AZURE_USER_CLIENT_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckUserDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckUser(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "username": "user_object_id_import", "object_id": clientId, "type": "E", "roles": "[\"db_owner\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("mssql_user.test_import"),
				),
			},
			{
				ResourceName:      "mssql_user.test_import",
				ImportState:       true,
				ImportStateIdFunc: testAccImportStateId("mssql_user.test_import", "not-a-guid", "azure=true"),
				ExpectError:       regexp.MustCompile(`expected "object_id" to be a valid UUID`),
			},
			{
				ResourceName:      "mssql_user.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_user.test_import", clientId, "azure=true"),
			},
		},
	})
}

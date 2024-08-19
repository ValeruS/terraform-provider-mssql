package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseCredential_Azure_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseCredemtialDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseCredential(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred_import", "identity_name": "test_identity_name_import", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.test_import"),
				),
			},
			{
				ResourceName:            "mssql_database_credential.test_import",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
				ImportStateIdFunc:       testAccImportStateId("mssql_database_credential.test_import", true),
			},
		},
	})
}

package mssql

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAzureExternalDatasource_Azure_BasicImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckAzureExternalDatasourceDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAzureExternalDatasource(t, "test_import", "azure", map[string]interface{}{"database": "testdb", "data_source_name": "test_datasource", "location": "fakesqlsrv1.database.windows.net", "type": "RDBMS", "remote_database_name": "test_db_remote", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureExternalDatasourceExists("mssql_azure_external_datasource.test_import"),
				),
			},
			{
				ResourceName:      "mssql_azure_external_datasource.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccImportStateId("mssql_azure_external_datasource.test_import", true),
			},
		},
	})
}

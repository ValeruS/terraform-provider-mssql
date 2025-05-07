package mssql

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccEntraIDLoginImportStateId(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found: %s", resource)
		}

		host := rs.Primary.Attributes["server.0.host"]
		port := "1433" // Default port
		if portAttr, ok := rs.Primary.Attributes["server.0.port"]; ok {
			port = portAttr
		}
		loginName := rs.Primary.Attributes["login_name"]

		return fmt.Sprintf("sqlserver://%s:%s/login/%s", host, port, loginName), nil
	}
}

func TestAccEntraIDLogin_Local_BasicImport(t *testing.T) {
	// Create mock connector
	mockConnector := &MockEntraIDLoginConnector{}

	// Create test provider with mock connector
	testProvider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"mssql_entraid_login": resourceEntraIDLogin(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return mockConnector, nil
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"mssql": func() (*schema.Provider, error) {
				return testProvider, nil
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			return testAccCheckEntraIDLoginDestroyMock(state, mockConnector)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "test_import", "azure", map[string]interface{}{
					"login_name": "test_import@example.com",
					"server": map[string]interface{}{
						"host": "test-server.database.windows.net",
						"azure_login": map[string]interface{}{
							"tenant_id": "mock-tenant-id",
							"client_id": "mock-client-id",
							"client_secret": "mock-client-secret",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExistsMock("mssql_entraid_login.test_import", mockConnector),
				),
			},
			{
				ResourceName:      "mssql_entraid_login.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEntraIDLoginImportStateId("mssql_entraid_login.test_import"),
				ImportStateVerifyIgnore: []string{
					"server.0.azure_login.#",
					"server.0.azure_login.0.%",
					"server.0.azure_login.0.client_id",
					"server.0.azure_login.0.client_secret",
					"server.0.azure_login.0.tenant_id",
					"server.0.login.#",
					"server.0.login.0.%",
					"server.0.login.0.password",
					"server.0.login.0.username",
					"object_id",
				},
			},
		},
	})
}

func TestAccEntraIDLogin_ObjectID_Import(t *testing.T) {
	// Create mock connector
	mockConnector := &MockEntraIDLoginConnector{}

	// Create test provider with mock connector
	testProvider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"mssql_entraid_login": resourceEntraIDLogin(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return mockConnector, nil
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"mssql": func() (*schema.Provider, error) {
				return testProvider, nil
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			return testAccCheckEntraIDLoginDestroyMock(state, mockConnector)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccCheckEntraIDLogin(t, "test_import_with_object_id", "azure", map[string]interface{}{
					"login_name": "test_import_with_object_id@example.com",
					"object_id": "12345678-1234-1234-1234-123456789012",
					"server": map[string]interface{}{
						"host": "test-server.database.windows.net",
						"azure_login": map[string]interface{}{
							"tenant_id": "mock-tenant-id",
							"client_id": "mock-client-id",
							"client_secret": "mock-client-secret",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExistsMock("mssql_entraid_login.test_import_with_object_id", mockConnector),
					resource.TestCheckResourceAttr("mssql_entraid_login.test_import_with_object_id", "object_id", "12345678-1234-1234-1234-123456789012"),
				),
			},
			{
				ResourceName:      "mssql_entraid_login.test_import_with_object_id",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEntraIDLoginImportStateId("mssql_entraid_login.test_import_with_object_id"),
				ImportStateVerifyIgnore: []string{
					"server.0.azure_login.#",
					"server.0.azure_login.0.%",
					"server.0.azure_login.0.client_id",
					"server.0.azure_login.0.client_secret",
					"server.0.azure_login.0.tenant_id",
					"server.0.login.#",
					"server.0.login.0.%",
					"server.0.login.0.password",
					"server.0.login.0.username",
				},
			},
		},
	})
}

package mssql

import (
	"context"
	"fmt"
	"testing"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/rs/zerolog"
)

// MockEntraIDLoginConnector implements EntraIDLoginConnector interface for testing
type MockEntraIDLoginConnector struct {
	login *model.EntraIDLogin
}

func (m *MockEntraIDLoginConnector) GetConnector(prefix string, data *schema.ResourceData) (interface{}, error) {
	return m, nil
}

func (m *MockEntraIDLoginConnector) ResourceLogger(resource, function string) zerolog.Logger {
	return zerolog.Nop()
}

func (m *MockEntraIDLoginConnector) DataSourceLogger(datasource, function string) zerolog.Logger {
	return zerolog.Nop()
}

func (m *MockEntraIDLoginConnector) CreateEntraIDLogin(ctx context.Context, name, objectId string) error {
	m.login = &model.EntraIDLogin{
		LoginName:       name,
		ObjectId:        objectId,
		PrincipalID:     1001, // Mock principal ID
		Sid:             "test-sid",
		DefaultDatabase: "master",
		DefaultLanguage: "us_english",
	}
	return nil
}

func (m *MockEntraIDLoginConnector) GetEntraIDLogin(ctx context.Context, name string) (*model.EntraIDLogin, error) {
	if m.login != nil && m.login.LoginName == name {
		return m.login, nil
	}
	return nil, nil
}

func (m *MockEntraIDLoginConnector) DeleteEntraIDLogin(ctx context.Context, name string) error {
	if m.login != nil && m.login.LoginName == name {
		m.login = nil
		return nil
	}
	return nil
}

func TestAccEntraIDLogin_Mock(t *testing.T) {
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
				Config: testAccCheckEntraIDLogin(t, "mock", "azure", map[string]interface{}{
					"login_name": "mock@example.com",
					"server": map[string]interface{}{
						"host": "localhost",
						"azure_login": map[string]interface{}{
							"tenant_id": "mock-tenant-id",
							"client_id": "mock-client-id",
							"client_secret": "mock-client-secret",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExistsMock("mssql_entraid_login.mock", mockConnector),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock", "login_name", "mock@example.com"),
					resource.TestCheckResourceAttrSet("mssql_entraid_login.mock", "principal_id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock", "server.0.azure_login.0.tenant_id", "mock-tenant-id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock", "server.0.azure_login.0.client_id", "mock-client-id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock", "server.0.azure_login.0.client_secret", "mock-client-secret"),
				),
			},
		},
	})
}

func TestAccEntraIDLogin_ObjectID_Mock(t *testing.T) {
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
				Config: testAccCheckEntraIDLogin(t, "mock_with_object_id", "azure", map[string]interface{}{
					"login_name": "mock@example.com",
					"object_id": "12345678-1234-1234-1234-123456789012",
					"server": map[string]interface{}{
						"host": "localhost",
						"azure_login": map[string]interface{}{
							"tenant_id": "mock-tenant-id",
							"client_id": "mock-client-id",
							"client_secret": "mock-client-secret",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEntraIDLoginExistsMock("mssql_entraid_login.mock_with_object_id", mockConnector),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "login_name", "mock@example.com"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "object_id", "12345678-1234-1234-1234-123456789012"),
					resource.TestCheckResourceAttrSet("mssql_entraid_login.mock_with_object_id", "principal_id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "server.0.azure_login.0.tenant_id", "mock-tenant-id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "server.0.azure_login.0.client_id", "mock-client-id"),
					resource.TestCheckResourceAttr("mssql_entraid_login.mock_with_object_id", "server.0.azure_login.0.client_secret", "mock-client-secret"),
				),
			},
		},
	})
}

func testAccCheckEntraIDLoginDestroyMock(state *terraform.State, mockConnector *MockEntraIDLoginConnector) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_entraid_login" {
			continue
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := mockConnector.GetEntraIDLogin(context.Background(), loginName)
		if login != nil {
			return fmt.Errorf("login still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckEntraIDLoginExistsMock(resource string, mockConnector *MockEntraIDLoginConnector) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_entraid_login" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_entraid_login", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := mockConnector.GetEntraIDLogin(context.Background(), loginName)
		if login == nil {
			return fmt.Errorf("login does not exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		// Verify object ID if it's set in the resource
		if objectId, ok := rs.Primary.Attributes["object_id"]; ok && objectId != "" {
			if login.ObjectId != objectId {
				return fmt.Errorf("expected object_id %s, got %s", objectId, login.ObjectId)
			}
		}

		return nil
	}
}

func testAccCheckEntraIDLogin(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `
			resource "mssql_entraid_login" "{{ .name }}" {
				server {
					host = "{{ .server.host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {
						tenant_id = "{{ .server.azure_login.tenant_id }}"
						client_id = "{{ .server.azure_login.client_id }}"
						client_secret = "{{ .server.azure_login.client_secret }}"
					}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				{{with .object_id}}object_id = "{{ . }}"{{end}}
			}`

	data["name"] = name
	data["login"] = login
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

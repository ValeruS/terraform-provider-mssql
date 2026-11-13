package mssql

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatabaseCredential_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseCredemtialDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseCredential(t, "test_credential", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.test_credential"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "identity_name", "test_identity_name"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_credential.test_credential", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.test_credential", "principal_id"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.test_credential", "credential_id"),
				),
			},
		},
	})
}

func TestAccDatabaseCredential_Azure_Basic_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseCredemtialDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseCredential(t, "update", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name_1", "secret": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.update"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "identity_name", "test_identity_name_1"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.update", "principal_id"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.update", "credential_id"),
				),
			},
			{
				Config: testAccCheckDatabaseCredential(t, "update", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred", "identity_name": "test_identity_name_2", "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.update"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "database", "testdb"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "credential_name", "test_scoped_cred"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "identity_name", "test_identity_name_2"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_database_credential.update", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.update", "principal_id"),
					resource.TestCheckResourceAttrSet("mssql_database_credential.update", "credential_id"),
				),
			},
		},
	})
}

func TestAccDatabaseCredential_Azure_SecretWo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckDatabaseCredemtialDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatabaseCredential(t, "wo", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred_wo", "identity_name": "test_identity_name_1", "secret_wo": "V3ryS3cretP@asswd", "secret_wo_version": 1, "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.wo"),
					resource.TestCheckResourceAttr("mssql_database_credential.wo", "identity_name", "test_identity_name_1"),
					resource.TestCheckNoResourceAttr("mssql_database_credential.wo", "secret"),
					resource.TestCheckNoResourceAttr("mssql_database_credential.wo", "secret_wo"),
					resource.TestCheckResourceAttr("mssql_database_credential.wo", "secret_wo_version", "1"),
				),
			},
			{
				// Changing identity_name without bumping the version keeps the existing secret
				Config: testAccCheckDatabaseCredential(t, "wo", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred_wo", "identity_name": "test_identity_name_2", "secret_wo": "V3ryS3cretP@asswd", "secret_wo_version": 1, "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.wo"),
					resource.TestCheckResourceAttr("mssql_database_credential.wo", "identity_name", "test_identity_name_2"),
					resource.TestCheckNoResourceAttr("mssql_database_credential.wo", "secret_wo"),
				),
			},
			{
				Config: testAccCheckDatabaseCredential(t, "wo", "azure", map[string]interface{}{"database": "testdb", "credential_name": "test_scoped_cred_wo", "identity_name": "test_identity_name_2", "secret_wo": "V3ryS3cretP@asswdUpdated123", "secret_wo_version": 2, "password": "V3ryS3cretP@asswd!Key"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseCredemtialExists("mssql_database_credential.wo"),
					resource.TestCheckNoResourceAttr("mssql_database_credential.wo", "secret"),
					resource.TestCheckNoResourceAttr("mssql_database_credential.wo", "secret_wo"),
					resource.TestCheckResourceAttr("mssql_database_credential.wo", "secret_wo_version", "2"),
				),
			},
		},
	})
}

func TestAccDatabaseCredential_Local_SecretWo_Validate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatabaseCredential(t, "wo_validate", "login", map[string]interface{}{"database": "master", "credential_name": "wo_validate", "identity_name": "wo_validate", "secret": "V3ryS3cretP@asswd", "secret_wo": "V3ryS3cretP@asswd", "secret_wo_version": 1, "password": "V3ryS3cretP@asswd!Key"}),
				ExpectError: regexp.MustCompile(`"secret_wo": conflicts with secret`),
			},
			{
				Config:      testAccCheckDatabaseCredential(t, "wo_validate", "login", map[string]interface{}{"database": "master", "credential_name": "wo_validate", "identity_name": "wo_validate", "secret_wo": "V3ryS3cretP@asswd", "password": "V3ryS3cretP@asswd!Key"}),
				ExpectError: regexp.MustCompile(`all of .secret_wo,secret_wo_version. must be\s+specified`),
			},
			{
				Config:      testAccCheckDatabaseCredential(t, "wo_validate", "login", map[string]interface{}{"database": "master", "credential_name": "wo_validate", "identity_name": "wo_validate", "secret_wo_version": 1, "password": "V3ryS3cretP@asswd!Key"}),
				ExpectError: regexp.MustCompile(`all of .secret_wo,secret_wo_version. must be\s+specified`),
			},
		},
	})
}

func testAccCheckDatabaseCredential(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_database_masterkey" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				password = "{{ .password }}"
			}
			resource "mssql_database_credential" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database = "{{ .database }}"
				credential_name = "{{ .credential_name }}"
				identity_name = "{{ .identity_name }}"
				{{ with .secret }}secret = "{{ . }}"{{ end }}
				{{ with .secret_wo }}secret_wo = "{{ . }}"{{ end }}
				{{ with .secret_wo_version }}secret_wo_version = {{ . }}{{ end }}
				depends_on = [mssql_database_masterkey.{{ .name }}]
			}`

	data["name"] = name
	data["login"] = login
	switch login {
	case "fedauth", "msi", "azure":
		data["host"] = os.Getenv("TF_ACC_SQL_SERVER")
	case "login":
		data["host"] = "localhost"
	default:
		t.Fatalf("login expected to be one of 'login', 'azure', 'msi', 'fedauth', got %s", login)
	}
	res, err := templateToString(name, text, data)
	if err != nil {
		t.Fatalf("%s", err)
	}
	return res
}

func testAccCheckDatabaseCredemtialDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_database_credential" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		database := rs.Primary.Attributes["database"]
		credentialname := rs.Primary.Attributes["credential_name"]
		scopedcredential, err := connector.GetDatabaseCredential(database, credentialname)
		if scopedcredential != nil {
			return fmt.Errorf("database scoped credential still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckDatabaseCredemtialExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_database_credential" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_database_credential", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		database := rs.Primary.Attributes["database"]
		credentialname := rs.Primary.Attributes["credential_name"]
		scopedcredential, err := connector.GetDatabaseCredential(database, credentialname)
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
		if scopedcredential.CredentialName != credentialname {
			return fmt.Errorf("expected to be credential_name %s, got %s", credentialname, scopedcredential.CredentialName)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "credential_name":
				actual = scopedcredential.CredentialName
			case "identity_name":
				actual = scopedcredential.IdentityName
			default:
				return fmt.Errorf("unknown property %s", check.name)
			}
			if (check.op == "" || check.op == "==") && !equal(check.expected, actual) {
				return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
			}
			if check.op == "!=" && equal(check.expected, actual) {
				return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
			}
		}
		return nil
	}
}

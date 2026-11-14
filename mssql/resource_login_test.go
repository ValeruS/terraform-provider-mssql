package mssql

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccLogin_Local_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "basic", "login", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.basic"),
					testAccCheckLoginWorks("mssql_login.basic", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccLogin_Local_Basic_Pass_Validate_Length(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckLogin(t, "basic_pass_validate", "login", map[string]interface{}{"login_name": "login_basic_pass_length", "password": "shotpas"}),
				ExpectError: regexp.MustCompile("length should equal to or greater than 8"),
			},
		},
	})
}

func TestAccLogin_Local_Basic_SID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "basic", "login", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡", "sid": "0xB7BDEF7990D03541BAA2AD73E4FF18E8"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.basic"),
					testAccCheckLoginWorks("mssql_login.basic", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "sid", "0xB7BDEF7990D03541BAA2AD73E4FF18E8"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccLogin_Azure_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "basic", "azure", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccLogin_Azure_Basic_SID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "basic", "azure", map[string]interface{}{"login_name": "login_basic", "password": "valueIsH8kd$¡", "sid": "0x01060000000000640000000000000000BAF5FC800B97EF49AC6FD89469C4987F"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "login_name", "login_basic"),
					resource.TestCheckResourceAttr("mssql_login.basic", "password", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.basic", "sid", "0x01060000000000640000000000000000BAF5FC800B97EF49AC6FD89469C4987F"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_database", "master"),
					resource.TestCheckResourceAttr("mssql_login.basic", "default_language", "us_english"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_login.basic", "server.0.login.#", "0"),
					resource.TestCheckResourceAttrSet("mssql_login.basic", "principal_id"),
				),
			},
		},
	})
}

func TestAccLogin_Local_UpdateLoginName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update_pre", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update_post", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_post"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
		}})
}

func TestAccLogin_Local_UpdatePassword(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "password", "valueIsH8kd$¡"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "otherIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "password", "otherIsH8kd$¡"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckLoginWorks("mssql_login.test_update", "otherIsH8kd$¡"),
				),
			},
		}})
}

func TestAccLogin_Local_UpdateDefaultDatabase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "master"),
					testAccCheckLoginExists("mssql_login.test_update", Check{"default_database", "==", "master"}),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡", "default_database": "tempdb"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "default_database", "tempdb"),
					testAccCheckLoginExists("mssql_login.test_update", Check{"default_database", "==", "tempdb"}),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
		}})
}

func TestAccLogin_Local_UpdateDefaultLanguage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "us_english"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "login", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡", "default_language": "russian"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "default_language", "russian"),
					testAccCheckLoginExists("mssql_login.test_update", Check{"default_language", "==", "russian"}),
					testAccCheckLoginWorks("mssql_login.test_update", "valueIsH8kd$¡"),
				),
			},
		}})
}

func TestAccLogin_Azure_UpdateLoginName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_update", "azure", map[string]interface{}{"login_name": "login_update_pre", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_pre"),
					testAccCheckLoginExists("mssql_login.test_update"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "azure", map[string]interface{}{"login_name": "login_update_post", "password": "valueIsH8kd$¡"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "login_name", "login_update_post"),
					testAccCheckLoginExists("mssql_login.test_update"),
				),
			},
		}})
}

func TestAccLogin_Azure_UpdatePassword(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				// On Azure SQL Database a login can only connect to a database where it has a user
				Config: testAccCheckLogin(t, "test_update", "azure", map[string]interface{}{"login_name": "login_update", "password": "valueIsH8kd$¡", "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "password", "valueIsH8kd$¡"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckDatabaseUserWorks("mssql_user.test_update", "login_update", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_update", "azure", map[string]interface{}{"login_name": "login_update", "password": "otherIsH8kd$¡", "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_login.test_update", "password", "otherIsH8kd$¡"),
					testAccCheckLoginExists("mssql_login.test_update"),
					testAccCheckDatabaseUserWorks("mssql_user.test_update", "login_update", "otherIsH8kd$¡"),
				),
			},
		}})
}

func TestAccLogin_Azure_PasswordWo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				// On Azure SQL Database a login can only connect to a database where it has a user
				Config: testAccCheckLogin(t, "test_wo", "azure", map[string]interface{}{"login_name": "login_wo", "password_wo": "valueIsH8kd$¡", "password_wo_version": 1, "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.test_wo"),
					testAccCheckDatabaseUserWorks("mssql_user.test_wo", "login_wo", "valueIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password_wo_version", "1"),
					resource.TestCheckResourceAttrSet("mssql_login.test_wo", "principal_id"),
				),
			},
			{
				// Changing the value without bumping the version keeps the password
				Config: testAccCheckLogin(t, "test_wo", "azure", map[string]interface{}{"login_name": "login_wo", "password_wo": "otherIsH8kd$¡", "password_wo_version": 1, "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseUserWorks("mssql_user.test_wo", "login_wo", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_wo", "azure", map[string]interface{}{"login_name": "login_wo", "password_wo": "otherIsH8kd$¡", "password_wo_version": 2, "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.test_wo"),
					testAccCheckDatabaseUserWorks("mssql_user.test_wo", "login_wo", "otherIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password_wo_version", "2"),
				),
			},
			{
				// Switching back to the stored password
				Config: testAccCheckLogin(t, "test_wo", "azure", map[string]interface{}{"login_name": "login_wo", "password": "thirdIsH8kd$¡", "user_database": "testdb"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.test_wo"),
					testAccCheckDatabaseUserWorks("mssql_user.test_wo", "login_wo", "thirdIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password", "thirdIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
				),
			},
		},
	})
}

func TestAccLogin_Local_PasswordWo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckLoginDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password_wo": "valueIsH8kd$¡", "password_wo_version": 1}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginExists("mssql_login.test_wo"),
					testAccCheckLoginWorks("mssql_login.test_wo", "valueIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password_wo_version", "1"),
				),
			},
			{
				// Changing default_language without bumping the version keeps the password
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password_wo": "valueIsH8kd$¡", "password_wo_version": 1, "default_language": "russian"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginWorks("mssql_login.test_wo", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "default_language", "russian"),
				),
			},
			{
				// Changing the value without bumping the version keeps the password,
				// even when another argument triggers an update
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password_wo": "otherIsH8kd$¡", "password_wo_version": 1, "default_language": "us_english"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginWorks("mssql_login.test_wo", "valueIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "default_language", "us_english"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password_wo": "otherIsH8kd$¡", "password_wo_version": 1, "default_language": "russian"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginWorks("mssql_login.test_wo", "valueIsH8kd$¡"),
				),
			},
			{
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password_wo": "otherIsH8kd$¡", "password_wo_version": 2, "default_language": "russian"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginWorks("mssql_login.test_wo", "otherIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password_wo_version", "2"),
				),
			},
			{
				// Switching back to the stored password
				Config: testAccCheckLogin(t, "test_wo", "login", map[string]interface{}{"login_name": "login_wo", "password": "thirdIsH8kd$¡", "default_language": "russian"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoginWorks("mssql_login.test_wo", "thirdIsH8kd$¡"),
					resource.TestCheckResourceAttr("mssql_login.test_wo", "password", "thirdIsH8kd$¡"),
					resource.TestCheckNoResourceAttr("mssql_login.test_wo", "password_wo"),
				),
			},
		},
	})
}

func TestAccLogin_Local_PasswordWo_Validate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckLogin(t, "wo_validate", "login", map[string]interface{}{"login_name": "login_wo_validate"}),
				ExpectError: regexp.MustCompile(`one of .password,password_wo. must be\s+specified`),
			},
			{
				Config:      testAccCheckLogin(t, "wo_validate", "login", map[string]interface{}{"login_name": "login_wo_validate", "password": "valueIsH8kd$¡", "password_wo": "valueIsH8kd$¡", "password_wo_version": 1}),
				ExpectError: regexp.MustCompile(`"password_wo": conflicts with password`),
			},
			{
				Config:      testAccCheckLogin(t, "wo_validate", "login", map[string]interface{}{"login_name": "login_wo_validate", "password_wo": "valueIsH8kd$¡"}),
				ExpectError: regexp.MustCompile(`all of .password_wo,password_wo_version. must be\s+specified`),
			},
			{
				Config:      testAccCheckLogin(t, "wo_validate", "login", map[string]interface{}{"login_name": "login_wo_validate", "password_wo": "shotpas", "password_wo_version": 1}),
				ExpectError: regexp.MustCompile(`.*password_wo.*`),
			},
		},
	})
}

func testAccCheckLogin(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_login" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "{{ .login_name }}"
				{{ with .password }}password = "{{ . }}"{{ end }}
				{{ with .password_wo }}password_wo = "{{ . }}"{{ end }}
				{{ with .password_wo_version }}password_wo_version = {{ . }}{{ end }}
				{{ with .sid }}sid = "{{ . }}"{{ end }}
				{{ with .default_database }}default_database = "{{ . }}"{{ end }}
				{{ with .default_language }}default_language = "{{ . }}"{{ end }}
			}
			{{ with .user_database }}
			resource "mssql_user" "{{ $.name }}" {
				server {
					host = "{{ $.host }}"
					{{if eq $.login "fedauth"}}azuread_default_chain_auth {}{{ else if eq $.login "msi"}}azuread_managed_identity_auth {}{{ else if eq $.login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				database   = "{{ . }}"
				username   = "{{ $.login_name }}"
				login_name = mssql_login.{{ $.name }}.login_name
			}
			{{ end }}`

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

func testAccCheckLoginDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_login" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := connector.GetLogin(loginName)
		if login != nil {
			return fmt.Errorf("login still exists")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
	}
	return nil
}

func testAccCheckLoginExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_login" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_login", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		loginName := rs.Primary.Attributes["login_name"]
		login, err := connector.GetLogin(loginName)
		if login == nil {
			return fmt.Errorf("login does not exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "default_database":
				actual = login.DefaultDatabase
			case "default_language":
				actual = login.DefaultLanguage
			default:
				return fmt.Errorf("unknown property %s", check.name)
			}
			if (check.op == "" || check.op == "==") && check.expected != actual {
				return fmt.Errorf("expected %s == %s, got %s", check.name, check.expected, actual)
			}
			if check.op == "!=" && check.expected == actual {
				return fmt.Errorf("expected %s != %s, got %s", check.name, check.expected, actual)
			}
		}
		return nil
	}
}

func testAccCheckLoginWorks(resource string, password string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_login" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_login", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestLoginConnector(rs.Primary.Attributes, password)
		if err != nil {
			return err
		}
		systemUser, err := connector.GetSystemUser()
		if err != nil {
			return err
		}
		if systemUser != rs.Primary.Attributes[loginNameProp] {
			return fmt.Errorf("expected to log in as [%s], got [%s]", rs.Primary.Attributes[loginNameProp], systemUser)
		}
		return nil
	}
}

package mssql

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServerRoleMember_Local_Basic_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "local_test_create", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_create"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_create", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Local_Basic_Update_Add(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "local_test_add", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_add"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.azure_login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "local_test_add", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_1\", \"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_add"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "members.1", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_add", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Local_Basic_Update_Drop(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "local_test_drop", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_1\", \"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_drop"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "members.1", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.azure_login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "local_test_drop", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_drop"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "members.0", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_drop", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Local_Basic_Update_Both(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		IsUnitTest:        runLocalAccTests,
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "local_test_both", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_0\", \"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_both"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.0", "login_test_0"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.1", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.azure_login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "local_test_both", "login", map[string]interface{}{"role_name": "sysadmin", "members": "[\"login_test_2\", \"login_test_3\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.local_test_both"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "role_name", "sysadmin"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.0", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "members.1", "login_test_3"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.host", "localhost"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.0.username", os.Getenv("MSSQL_USERNAME")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.login.0.password", os.Getenv("MSSQL_PASSWORD")),
					resource.TestCheckResourceAttr("mssql_server_role_member.local_test_both", "server.0.azure_login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Azure_Basic_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_create", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_create"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_create", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Azure_Basic_Update_Add(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_add", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_add"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_add", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_1\", \"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_add"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "members.1", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_add", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Azure_Basic_Update_Drop(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_drop", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_1\", \"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_drop"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "members.0", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "members.1", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_drop", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_2\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_drop"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "members.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "members.0", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_drop", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func TestAccServerRoleMember_Azure_Basic_Update_Both(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      func(state *terraform.State) error { return testAccCheckServerRoleMemberDestroy(state) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_both", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_0\", \"login_test_1\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_both"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.0", "login_test_0"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.1", "login_test_1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.login.#", "0"),
				),
			},
			{
				Config: testAccCheckServerRoleMember(t, "azure_test_both", "azure", map[string]interface{}{"role_name": "##MS_LoginManager##", "members": "[\"login_test_2\", \"login_test_3\"]"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerRoleMemberExists("mssql_server_role_member.azure_test_both"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "role_name", "##MS_LoginManager##"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.#", "2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.0", "login_test_2"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "members.1", "login_test_3"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.host", os.Getenv("TF_ACC_SQL_SERVER")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.port", "1433"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.#", "1"),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.tenant_id", os.Getenv("MSSQL_TENANT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.client_id", os.Getenv("MSSQL_CLIENT_ID")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.azure_login.0.client_secret", os.Getenv("MSSQL_CLIENT_SECRET")),
					resource.TestCheckResourceAttr("mssql_server_role_member.azure_test_both", "server.0.login.#", "0"),
				),
			},
		},
	})
}

func testAccCheckServerRoleMember(t *testing.T, name string, login string, data map[string]interface{}) string {
	text := `resource "mssql_login" "{{ .name }}" {
				count = 4
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				login_name = "login_test_${count.index}"
				password   = "valueIsH8kd$A"
			}
			resource "mssql_server_role_member" "{{ .name }}" {
				server {
					host = "{{ .host }}"
					{{if eq .login "fedauth"}}azuread_default_chain_auth {}{{ else if eq .login "msi"}}azuread_managed_identity_auth {}{{ else if eq .login "azure" }}azure_login {}{{ else }}login {}{{ end }}
				}
				role_name = "{{ .role_name }}"
				members = {{ .members }}
				depends_on = [mssql_login.{{ .name }}]
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

func getManagedMembersFromState(attrs map[string]string) []string {
	count, _ := strconv.Atoi(attrs["members.#"])
	if count <= 0 {
		return nil
	}
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if v := attrs[fmt.Sprintf("members.%d", i)]; v != "" {
			out = append(out, v)
		}
	}
	return out
}

func testAccCheckServerRoleMemberDestroy(state *terraform.State) error {
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "mssql_server_role_member" {
			continue
		}

		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		roleName := rs.Primary.Attributes["role_name"]
		managedMembers := getManagedMembersFromState(rs.Primary.Attributes)
		roleMembers, err := connector.GetServerRoleMember(roleName, managedMembers)
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}
		if roleMembers != nil && len(roleMembers.Members) > 0 {
			return fmt.Errorf("role members still exist")
		}
	}
	return nil
}

func testAccCheckServerRoleMemberExists(resource string, checks ...Check) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Type != "mssql_server_role_member" {
			return fmt.Errorf("expected resource of type %s, got %s", "mssql_server_role_member", rs.Type)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}
		connector, err := getTestConnector(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		roleName := rs.Primary.Attributes["role_name"]
		managedMembers := getManagedMembersFromState(rs.Primary.Attributes)
		roleMembers, err := connector.GetServerRoleMember(roleName, managedMembers)
		if roleMembers == nil {
			return fmt.Errorf("role members do not exist")
		}
		if err != nil {
			return fmt.Errorf("expected no error, got %s", err)
		}

		var actual interface{}
		for _, check := range checks {
			switch check.name {
			case "role_name":
				actual = roleMembers.RoleName
			case "members":
				actual = roleMembers.Members
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

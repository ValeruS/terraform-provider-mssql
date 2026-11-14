package mssql

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// A change of the credentials in the server block, for example a rotated
// MSSQL_PASSWORD, must be an in-place update: it must not replace the resource,
// and the resource must support update so the new credentials can be saved.
func TestServerBlockCredentialsChangeIsInPlace(t *testing.T) {
	// Values for attributes whose CustomizeDiff checks their format.
	overrides := map[string]string{
		sqlscriptProp:    "Q1JFQVRFIFRBQkxFIFVzZXJzIChpZCBpbnQp", // base64 of "CREATE TABLE Users (id int)"
		verifyObjectProp: "TABLE Users",
	}
	for name, r := range Provider(nil).ResourcesMap {
		if _, ok := r.Schema[serverProp]; !ok {
			continue
		}
		t.Run(name, func(t *testing.T) {
			if r.UpdateContext == nil {
				t.Fatal("resource has a server block but no update")
			}

			attributes := map[string]string{"id": "x"}
			config := map[string]interface{}{}
			// Set the required top-level attributes in state and config, so that the
			// diff can be computed. Only the server attribute is checked below.
			for k, s := range r.Schema {
				if k == serverProp || !s.Required {
					continue
				}
				switch s.Type {
				case schema.TypeString:
					value := "value"
					if v, ok := overrides[k]; ok {
						value = v
					}
					attributes[k] = value
					config[k] = value
				case schema.TypeInt:
					attributes[k] = "1"
					config[k] = 1
				case schema.TypeSet, schema.TypeList:
					attributes[k+".#"] = "0"
					config[k] = []interface{}{}
				}
			}
			for k, v := range map[string]string{
				"server.#":                                 "1",
				"server.0.host":                            "example.database.windows.net",
				"server.0.port":                            "1433",
				"server.0.login.#":                         "1",
				"server.0.login.0.username":                "sa",
				"server.0.login.0.password":                "old",
				"server.0.azure_login.#":                   "0",
				"server.0.azuread_default_chain_auth.#":    "0",
				"server.0.azuread_managed_identity_auth.#": "0",
			} {
				attributes[k] = v
			}
			config[serverProp] = []interface{}{map[string]interface{}{
				"host":  "example.database.windows.net",
				"login": []interface{}{map[string]interface{}{"username": "sa", "password": "new"}},
			}}

			state := &terraform.InstanceState{ID: "x", Attributes: attributes}
			diff, err := r.SimpleDiff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatalf("diff returned error: %v", err)
			}
			attr, ok := diff.Attributes["server.0.login.0.password"]
			if !ok {
				t.Fatal("expected a diff on server.0.login.0.password")
			}
			if attr.RequiresNew {
				t.Error("a change of the server credentials should not replace the resource")
			}
		})
	}
}

package mssql

import (
	"encoding/base64"
	"fmt"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rs/zerolog"
)

func getLoginID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	loginName := data.Get(loginNameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/login/%s", host, port, loginName)
}

func getUserID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/user/%s", host, port, database, username)
}

func getDatabasePermissionsID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	username := data.Get(usernameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/permission/%s", host, port, database, username)
}

func getDatabaseRoleID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	roleName := data.Get(roleNameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/role/%s", host, port, database, roleName)
}

func getDatabaseSchemaID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	schemaName := data.Get(schemaNameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/schema/%s", host, port, database, schemaName)
}

func getDatabaseCredentialID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	credentialname := data.Get(credentialNameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/credential/%s", host, port, database, credentialname)
}

func getDatabaseMasterkeyID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/masterkey", host, port, database)
}

func getAzureExternalDatasourceID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	datasourcename := data.Get(datasourcenameProp).(string)
	return fmt.Sprintf("sqlserver://%s:%s/%s/externaldatasource/%s", host, port, database, datasourcename)
}

func getDatabaseSQLScriptID(data *schema.ResourceData) string {
	host := data.Get(serverProp + ".0.host").(string)
	port := data.Get(serverProp + ".0.port").(string)
	database := data.Get(databaseProp).(string)
	verifyObject := data.Get(verifyObjectProp).(string)
	id := fmt.Sprintf("%s:%s", database, verifyObject)
	encodedID := base64.URLEncoding.EncodeToString([]byte(id))
	return fmt.Sprintf("sqlserver://%s:%s/%s/sqlscript/%s", host, port, database, encodedID)
}

func loggerFromMeta(meta interface{}, resource, function string) zerolog.Logger {
	return meta.(model.Provider).ResourceLogger(resource, function)
}

func toStringSlice(values []interface{}) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = v.(string)
	}
	return result
}

func equal(a, b interface{}) bool {
	switch a.(type) {
	case []string:
		aa := a.([]string)
		bb := b.([]string)
		if len(aa) != len(bb) {
			return false
		}
		for i, v := range aa {
			if v != bb[i] {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

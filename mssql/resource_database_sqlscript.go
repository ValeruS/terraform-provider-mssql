package mssql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/pkg/errors"
)

func resourceDatabaseSQLScript() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabaseSQLScriptCreate,
		ReadContext:   resourceDatabaseSQLScriptRead,
		UpdateContext: resourceDatabaseSQLScriptUpdate,
		DeleteContext: resourceDatabaseSQLScriptDelete,

		Schema: map[string]*schema.Schema{
			serverProp: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: getServerSchema(serverProp),
				},
			},
			databaseProp: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			sqlscriptProp: {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsBase64,
			},
			verifyObjectProp: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Object to verify existence (format: 'TYPE NAME' e.g., 'TABLE Users')",
				ValidateFunc: func(i interface{}, k string) ([]string, []error) {
					v := i.(string)
					if v == "" {
						return nil, []error{fmt.Errorf("verify_object cannot be empty")}
					}
					parts := strings.Fields(v)
					if len(parts) != 2 {
						return nil, []error{fmt.Errorf("verify_object must be in format 'TYPE NAME', got: %s", v)}
					}
					return nil, nil
				},
			},
		},
		CustomizeDiff: func(ctx context.Context, data *schema.ResourceDiff, i interface{}) error {
			// Get verify_object value
			verifyObject := data.Get(verifyObjectProp).(string)
			if verifyObject == "" {
				return nil
			}

			// Get the sqlscript content and decode from base64
			scriptBase64 := data.Get(sqlscriptProp).(string)
			scriptContent, err := base64.StdEncoding.DecodeString(scriptBase64)
			if err != nil {
				return errors.Wrap(err, "failed to decode base64 sqlscript")
			}

			// Parse verify_object to get type and name
			parts := strings.Fields(verifyObject)
			if len(parts) != 2 {
				return fmt.Errorf("verify_object must be in format 'TYPE NAME', got: %s", verifyObject)
			}

			objectType := strings.ToUpper(parts[0])
			objectName := parts[1]

			// Handle different object name formats
			quotedName := fmt.Sprintf("'%s'", objectName)
			bracketName := fmt.Sprintf("[%s]", objectName)
			doubleName := fmt.Sprintf("\"%s\"", objectName)

			// Convert script to lowercase for case-insensitive matching
			scriptLower := strings.ToLower(string(scriptContent))
			objectTypeLower := strings.ToLower(objectType)
			objectNameLower := strings.ToLower(objectName)
			quotedNameLower := strings.ToLower(quotedName)
			bracketNameLower := strings.ToLower(bracketName)
			doubleNameLower := strings.ToLower(doubleName)

			// Define patterns to look for
			patterns := []string{
				// Basic patterns
				fmt.Sprintf("create %s %s", objectTypeLower, objectNameLower),
				fmt.Sprintf("create %s %s", objectTypeLower, quotedNameLower),
				fmt.Sprintf("create %s %s", objectTypeLower, bracketNameLower),
				fmt.Sprintf("create %s %s", objectTypeLower, doubleNameLower),

				// ALTER patterns
				fmt.Sprintf("alter %s %s", objectTypeLower, objectNameLower),
				fmt.Sprintf("alter %s %s", objectTypeLower, quotedNameLower),
				fmt.Sprintf("alter %s %s", objectTypeLower, bracketNameLower),
				fmt.Sprintf("alter %s %s", objectTypeLower, doubleNameLower),

				// CREATE OR ALTER patterns
				fmt.Sprintf("create or alter %s %s", objectTypeLower, objectNameLower),
				fmt.Sprintf("create or alter %s %s", objectTypeLower, quotedNameLower),
				fmt.Sprintf("create or alter %s %s", objectTypeLower, bracketNameLower),
				fmt.Sprintf("create or alter %s %s", objectTypeLower, doubleNameLower),

				// DROP and CREATE patterns
				fmt.Sprintf("drop %s %s.*create %s %s", objectTypeLower, objectNameLower, objectTypeLower, objectNameLower),
				fmt.Sprintf("drop %s %s.*create %s %s", objectTypeLower, quotedNameLower, objectTypeLower, quotedNameLower),
				fmt.Sprintf("drop %s %s.*create %s %s", objectTypeLower, bracketNameLower, objectTypeLower, bracketNameLower),
				fmt.Sprintf("drop %s %s.*create %s %s", objectTypeLower, doubleNameLower, objectTypeLower, doubleNameLower),
			}

			found := false
			for _, pattern := range patterns {
				if strings.Contains(scriptLower, pattern) {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("verify_object '%s %s' is specified but no matching 'CREATE', 'ALTER', or 'CREATE OR ALTER' statement for this object was found in the script", objectType, objectName)
			}

			return nil
		},
		Timeouts: &schema.ResourceTimeout{
			Create: defaultTimeout,
			Read:   defaultTimeout,
			Update: defaultTimeout,
			Delete: defaultTimeout,
		},
	}
}

type DatabaseSQLScriptConnector interface {
	DataBaseExecuteScript(ctx context.Context, database string, script string) error
	DatabaseExists(ctx context.Context, database string) (bool, error)
}

// getScript retrieves the SQL script content from the script attribute and decodes it from base64
func getScript(data *schema.ResourceData) (string, error) {
	script := data.Get(sqlscriptProp).(string)
	decoded, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode base64 sqlscript")
	}
	return string(decoded), nil
}

// getObjectExistsQuery generates a SQL query to check if an object exists
func getObjectExistsQuery(objectSpec string) (string, error) {
	if objectSpec == "" {
		return "SELECT 1", nil
	}

	parts := strings.Fields(objectSpec)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid object specification: %s", objectSpec)
	}

	objectType := strings.ToUpper(parts[0])
	objectName := parts[1]

	// Split object name into schema and name parts if it contains a dot
	var schema_Name, object_Name string
	if nameParts := strings.Split(objectName, "."); len(nameParts) == 2 {
		schema_Name = nameParts[0]
		object_Name = nameParts[1]
	} else {
		// If no schema specified, use the object name as is
		object_Name = objectName
	}

	switch objectType {
	case "TABLE":
		if schema_Name != "" {
			return fmt.Sprintf(`
			SELECT 1 
			FROM sys.tables t 
			INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
			WHERE t.name = N'%s' AND s.name = N'%s'
		`, object_Name, schema_Name), nil
		}
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.tables t 
			WHERE t.name = N'%s'
		`, object_Name), nil
	case "VIEW":
		if schema_Name != "" {
			return fmt.Sprintf(`
			SELECT 1 
			FROM sys.views v
			INNER JOIN sys.schemas s ON v.schema_id = s.schema_id
			WHERE v.name = N'%s' AND s.name = N'%s'
		`, object_Name, schema_Name), nil
		}
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.views v 
			WHERE v.name = N'%s'
		`, object_Name), nil
	case "PROCEDURE", "PROC":
		if schema_Name != "" {
			return fmt.Sprintf(`
			SELECT 1 
			FROM sys.procedures p
			INNER JOIN sys.schemas s ON p.schema_id = s.schema_id
			WHERE p.name = N'%s' AND s.name = N'%s'
		`, object_Name, schema_Name), nil
		}
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.procedures p 
			WHERE p.name = N'%s'
		`, object_Name), nil
	case "FUNCTION", "FUNC":
		if schema_Name != "" {
			return fmt.Sprintf(`
			SELECT 1 
			FROM sys.objects o
			INNER JOIN sys.schemas s ON o.schema_id = s.schema_id
			WHERE o.type IN ('FN', 'IF', 'TF') 
			AND o.name = N'%s' AND s.name = N'%s'
		`, object_Name, schema_Name), nil
		}
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.objects o 
			WHERE o.type IN ('FN', 'IF', 'TF') 
			AND o.name = N'%s'
		`, object_Name), nil
	case "SCHEMA":
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.schemas s 
			WHERE s.name = N'%s'
		`, object_Name), nil
	case "TRIGGER", "TRG":
		if schema_Name != "" {
			return fmt.Sprintf(`
			SELECT 1 
			FROM sys.triggers t
			INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
			WHERE t.name = N'%s' AND s.name = N'%s'
		`, object_Name, schema_Name), nil
		}
		return fmt.Sprintf(`
			SELECT 1 
			FROM sys.triggers t 
			WHERE t.name = N'%s'
		`, object_Name), nil
	default:
		return "", fmt.Errorf("unsupported object type: %s", objectType)
	}
}

func resourceDatabaseSQLScriptCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "sqlscript", "create")
	logger.Debug().Msgf("Create %s", getDatabaseSQLScriptID(data))

	database := data.Get(databaseProp).(string)
	script, err := getScript(data)
	if err != nil {
		return diag.FromErr(err)
	}

	connector, err := getDatabaseSQLScriptConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := connector.DataBaseExecuteScript(ctx, database, script); err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to execute SQL script in database [%s]", database))
	}

	data.SetId(getDatabaseSQLScriptID(data))

	logger.Info().Msgf("executed SQL script in database [%s]", database)

	return resourceDatabaseSQLScriptRead(ctx, data, meta)
}

func resourceDatabaseSQLScriptRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "sqlscript", "read")
	logger.Debug().Msgf("Read %s", data.Id())

	database := data.Get(databaseProp).(string)
	verifyObject := data.Get(verifyObjectProp).(string)

	connector, err := getDatabaseSQLScriptConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	exists, err := connector.DatabaseExists(ctx, database)
	if err != nil {
		return diag.FromErr(errors.Wrapf(err, "unable to check if database [%s] exists", database))
	}
	if !exists {
		logger.Info().Msgf("Database [%s] does not exist", database)
		data.SetId("")
		return nil
	}

	// Generate the appropriate query based on verify_object
	query, err := getObjectExistsQuery(verifyObject)
	if err != nil {
		return diag.FromErr(err)
	}

	// Execute the verification query
	err = connector.DataBaseExecuteScript(ctx, database, query)
	if err != nil {
		// If we're verifying an object and it doesn't exist, mark the resource as gone
		if verifyObject != "" && (strings.Contains(err.Error(), "Invalid object name") ||
			strings.Contains(err.Error(), "does not exist") ||
			!strings.Contains(err.Error(), "affected")) {
			logger.Info().Msgf("Object [%s] in database [%s] does not exist, marking resource for recreation", verifyObject, database)
			data.SetId("")
			return nil
		}

		return diag.FromErr(errors.Wrapf(err, "unable to verify object in database [%s]", database))
	}

	return nil
}

func resourceDatabaseSQLScriptUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "sqlscript", "update")
	logger.Debug().Msgf("Update %s", data.Id())

	// Only run if script content has changed
	if !data.HasChange(sqlscriptProp) {
		return nil
	}

	database := data.Get(databaseProp).(string)
	script, err := getScript(data)
	if err != nil {
		return diag.FromErr(err)
	}

	connector, err := getDatabaseSQLScriptConnector(meta, data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Store the old script value in case we need to revert
	oldValue, _ := data.GetChange(sqlscriptProp)
	oldScript := oldValue.(string)

	if err := connector.DataBaseExecuteScript(ctx, database, script); err != nil {
		// If script execution fails, revert the state to the old script value
		if err := data.Set(sqlscriptProp, oldScript); err != nil {
			logger.Error().Err(err).Msg("Failed to revert sqlscript state after execution error")
		}
		return diag.FromErr(errors.Wrapf(err, "unable to execute SQL script in database [%s]", database))
	}

	data.SetId(getDatabaseSQLScriptID(data))

	logger.Info().Msgf("executed SQL script in database [%s]", database)

	return resourceDatabaseSQLScriptRead(ctx, data, meta)
}

func resourceDatabaseSQLScriptDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	logger := loggerFromMeta(meta, "sqlscript", "delete")
	logger.Debug().Msgf("Delete %s", data.Id())

	database := data.Get(databaseProp).(string)
	// Nothing to do on delete as the script has already been executed
	data.SetId("")

	logger.Info().Msgf("Nothing to do on delete as the script has already been executed in database [%s]", database)

	return nil
}

func getDatabaseSQLScriptConnector(meta interface{}, data *schema.ResourceData) (DatabaseSQLScriptConnector, error) {
	provider := meta.(model.Provider)
	connector, err := provider.GetConnector(serverProp, data)
	if err != nil {
		return nil, err
	}
	return connector.(DatabaseSQLScriptConnector), nil
}

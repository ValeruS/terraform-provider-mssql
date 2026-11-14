package sql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

// GetDatabaseSchemaPermission returns the permissions granted to a database principal on a schema, or nil when the principal or the schema does not exist.
func (c *Connector) GetDatabaseSchemaPermission(ctx context.Context, database, schemaName, principalName string) (*model.DatabaseSchemaPermission, error) {
	cmd := `SELECT pr.principal_id, s.schema_id, pe.permission_name
			FROM [sys].[database_principals] AS pr
			CROSS JOIN [sys].[schemas] AS s
			LEFT JOIN [sys].[database_permissions] AS pe
				ON pe.grantee_principal_id = pr.principal_id
				AND pe.class = 3
				AND pe.major_id = s.schema_id
				AND pe.[state] = 'G'
			WHERE pr.name = @principalName AND s.name = @schemaName`
	found := false
	permsModel := model.DatabaseSchemaPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		PrincipalName: principalName,
		Permissions:   make([]string, 0),
	}
	err := c.
		setDatabase(&database).
		QueryContext(ctx, cmd,
			func(r *sql.Rows) error {
				for r.Next() {
					var permissionName sql.NullString
					if err := r.Scan(&permsModel.PrincipalID, &permsModel.SchemaID, &permissionName); err != nil {
						return err
					}
					found = true
					if permissionName.Valid {
						permsModel.Permissions = append(permsModel.Permissions, permissionName.String)
					}
				}
				return nil
			},
			sql.Named("principalName", principalName),
			sql.Named("schemaName", schemaName),
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &permsModel, nil
}

func (c *Connector) CreateDatabaseSchemaPermission(ctx context.Context, permissions *model.DatabaseSchemaPermission) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
			DECLARE @permission_name nvarchar(max)
			OPEN perm_cur
			FETCH NEXT FROM perm_cur INTO @permission_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'GRANT ' + @permission_name + ' ON SCHEMA::' + QuoteName(@schemaName) + ' TO ' + QuoteName(@principalName)
					EXEC (@stmt)
					FETCH NEXT FROM perm_cur INTO @permission_name
				END
			CLOSE perm_cur
			DEALLOCATE perm_cur
			`
	return c.
		setDatabase(&permissions.DatabaseName).
		ExecContext(ctx, cmd,
			sql.Named("schemaName", permissions.SchemaName),
			sql.Named("principalName", permissions.PrincipalName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

func (c *Connector) UpdateDatabaseSchemaPermission(ctx context.Context, database, schemaName, principalName string, permissions []string, changeType string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE @applies nvarchar(10)
			IF @changeType = 'GRANT' SET @applies = 'TO'
			IF @changeType = 'REVOKE' SET @applies = 'FROM'
			IF @applies IS NULL THROW 51000, 'invalid change type', 1
			DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
			DECLARE @permission_name nvarchar(max)
			OPEN perm_cur
			FETCH NEXT FROM perm_cur INTO @permission_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = @changeType + ' ' + @permission_name + ' ON SCHEMA::' + QuoteName(@schemaName) + ' ' + @applies + ' ' + QuoteName(@principalName)
					EXEC (@stmt)
					FETCH NEXT FROM perm_cur INTO @permission_name
				END
			CLOSE perm_cur
			DEALLOCATE perm_cur
			`
	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("schemaName", schemaName),
			sql.Named("principalName", principalName),
			sql.Named("permissions", strings.Join(permissions, ",")),
			sql.Named("changeType", changeType),
		)
}

func (c *Connector) DeleteDatabaseSchemaPermission(ctx context.Context, permissions *model.DatabaseSchemaPermission) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
			DECLARE @permission_name nvarchar(max)
			OPEN perm_cur
			FETCH NEXT FROM perm_cur INTO @permission_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'REVOKE ' + @permission_name + ' ON SCHEMA::' + QuoteName(@schemaName) + ' FROM ' + QuoteName(@principalName)
					EXEC (@stmt)
					FETCH NEXT FROM perm_cur INTO @permission_name
				END
			CLOSE perm_cur
			DEALLOCATE perm_cur
			`
	return c.
		setDatabase(&permissions.DatabaseName).
		ExecContext(ctx, cmd,
			sql.Named("schemaName", permissions.SchemaName),
			sql.Named("principalName", permissions.PrincipalName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

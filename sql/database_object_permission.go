package sql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

// GetDatabaseObjectPermission returns the permissions granted to a database principal on an object, or on one of its
// columns when columnName is set, or nil when the principal, the object or the column does not exist.
func (c *Connector) GetDatabaseObjectPermission(ctx context.Context, database, schemaName, objectName, columnName, principalName string) (*model.DatabaseObjectPermission, error) {
	cmd := `SELECT pr.principal_id, o.object_id, pe.permission_name
			FROM [sys].[database_principals] AS pr
			CROSS JOIN [sys].[objects] AS o
			INNER JOIN [sys].[schemas] AS s
				ON o.schema_id = s.schema_id
			LEFT JOIN [sys].[columns] AS c
				ON c.object_id = o.object_id
				AND c.name = @columnName
			LEFT JOIN [sys].[database_permissions] AS pe
				ON pe.grantee_principal_id = pr.principal_id
				AND pe.class = 1
				AND pe.major_id = o.object_id
				AND pe.minor_id = ISNULL(c.column_id, 0)
				AND pe.[state] = 'G'
			WHERE pr.name = @principalName AND s.name = @schemaName AND o.name = @objectName
				AND (@columnName = '' OR c.column_id IS NOT NULL)`
	found := false
	permsModel := model.DatabaseObjectPermission{
		DatabaseName:  database,
		SchemaName:    schemaName,
		ObjectName:    objectName,
		ColumnName:    columnName,
		PrincipalName: principalName,
		Permissions:   make([]string, 0),
	}
	err := c.
		setDatabase(&database).
		QueryContext(ctx, cmd,
			func(r *sql.Rows) error {
				for r.Next() {
					var permissionName sql.NullString
					if err := r.Scan(&permsModel.PrincipalID, &permsModel.ObjectID, &permissionName); err != nil {
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
			sql.Named("objectName", objectName),
			sql.Named("columnName", columnName),
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

func (c *Connector) CreateDatabaseObjectPermission(ctx context.Context, permissions *model.DatabaseObjectPermission) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE @column nvarchar(max) = CASE WHEN @columnName = '' THEN '' ELSE ' (' + QuoteName(@columnName) + ')' END
			DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
			DECLARE @permission_name nvarchar(max)
			OPEN perm_cur
			FETCH NEXT FROM perm_cur INTO @permission_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'GRANT ' + @permission_name + ' ON OBJECT::' + QuoteName(@schemaName) + '.' + QuoteName(@objectName) + @column + ' TO ' + QuoteName(@principalName)
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
			sql.Named("objectName", permissions.ObjectName),
			sql.Named("columnName", permissions.ColumnName),
			sql.Named("principalName", permissions.PrincipalName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

func (c *Connector) UpdateDatabaseObjectPermission(ctx context.Context, database, schemaName, objectName, columnName, principalName string, permissions []string, changeType string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE @column nvarchar(max) = CASE WHEN @columnName = '' THEN '' ELSE ' (' + QuoteName(@columnName) + ')' END
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
					SET @stmt = @changeType + ' ' + @permission_name + ' ON OBJECT::' + QuoteName(@schemaName) + '.' + QuoteName(@objectName) + @column + ' ' + @applies + ' ' + QuoteName(@principalName)
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
			sql.Named("objectName", objectName),
			sql.Named("columnName", columnName),
			sql.Named("principalName", principalName),
			sql.Named("permissions", strings.Join(permissions, ",")),
			sql.Named("changeType", changeType),
		)
}

func (c *Connector) DeleteDatabaseObjectPermission(ctx context.Context, permissions *model.DatabaseObjectPermission) error {
	cmd := `DECLARE @stmt nvarchar(max)
			DECLARE @column nvarchar(max) = CASE WHEN @columnName = '' THEN '' ELSE ' (' + QuoteName(@columnName) + ')' END
			DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
			DECLARE @permission_name nvarchar(max)
			OPEN perm_cur
			FETCH NEXT FROM perm_cur INTO @permission_name
			WHILE @@FETCH_STATUS = 0
				BEGIN
					SET @stmt = 'REVOKE ' + @permission_name + ' ON OBJECT::' + QuoteName(@schemaName) + '.' + QuoteName(@objectName) + @column + ' FROM ' + QuoteName(@principalName)
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
			sql.Named("objectName", permissions.ObjectName),
			sql.Named("columnName", permissions.ColumnName),
			sql.Named("principalName", permissions.PrincipalName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

package sql

import (
	"context"
	"database/sql"
	"strings"

	"github.com/betr-io/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabasePermissions(ctx context.Context, database string, username string) (*model.DatabasePermissions, error) {
	cmd := `DECLARE @stmt nvarchar(max)
					SET @stmt = 'SELECT DISTINCT pr.principal_id, pr.name, ' +
											'pe.permission_name ' +
											'FROM [sys].[database_principals] AS pr LEFT JOIN [sys].[database_permissions] AS pe ' +
											'ON pe.grantee_principal_id = pr.principal_id ' +
											'WHERE pr.name = ' + QuoteName(@username, '''')
					EXEC (@stmt)`
	var (
		permissions []string
	)

	permsModel := model.DatabasePermissions{
		UserName:  username,
		DatabaseName: database,
		Permissions:  make([]string, 0),
	}

	err := c.
		setDatabase(&database).
		QueryContext(ctx, cmd,
			func(r *sql.Rows) error {
				for r.Next() {
					var name, permission_name string
					var principal_id string
					if err := r.Scan(&principal_id, &name, &permission_name); err != nil {
						// Check for a scan error.
						// Query rows will be closed with defer.
						return err
					}
					if permission_name == "CONNECT" {
						continue
					}
					permissions = append(permissions, permission_name)
				}
				return nil
			},
			sql.Named("database", database),
			sql.Named("username", username),
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(permissions) == 0 {
		permsModel.Permissions = make([]string, 0)
	} else {
		permsModel.Permissions = permissions
	}
	return &permsModel, nil
}

func (c *Connector) CreateDatabasePermissions(ctx context.Context, permissions *model.DatabasePermissions) error {
	cmd := `DECLARE @stmt nvarchar(max)
					DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
					DECLARE @permission_name nvarchar(max)
					OPEN perm_cur
					FETCH NEXT FROM perm_cur INTO @permission_name
					WHILE @@FETCH_STATUS = 0
						BEGIN
							SET @stmt = 'GRANT ' + @permission_name + ' TO ' + QuoteName(@username)
							EXEC (@stmt)
							FETCH NEXT FROM perm_cur INTO @permission_name
						END
					CLOSE perm_cur
					DEALLOCATE perm_cur
					`
	return c.
		setDatabase(&permissions.DatabaseName).
		ExecContext(ctx, cmd,
			sql.Named("username", permissions.UserName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

func (c *Connector) UpdateDatabasePermissions(ctx context.Context, permissions *model.DatabasePermissions) error {
	cmd := `DECLARE @stmt nvarchar(max)
					DECLARE grant_perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',') WHERE value NOT IN(SELECT permission_name FROM [sys].[database_permissions] pe, [sys].[database_principals] pr WHERE pe.grantee_principal_id = pr.principal_id AND pr.name = @username)
					DECLARE revoke_perm_cur CURSOR FOR SELECT pe.permission_name FROM [sys].[database_principals] pr LEFT JOIN [sys].[database_permissions] pe ON pe.grantee_principal_id = pr.principal_id AND pr.name = @username AND pe.permission_name != 'CONNECT' AND pe.permission_name NOT IN (SELECT value FROM String_Split(@permissions, ','))
					DECLARE @perm_name nvarchar(max)

					OPEN grant_perm_cur
					FETCH NEXT FROM grant_perm_cur INTO @perm_name
					WHILE @@FETCH_STATUS = 0
						BEGIN
							SET @stmt = 'GRANT ' + @perm_name + ' TO ' + QuoteName(@username)
							EXEC (@stmt)
							FETCH NEXT FROM grant_perm_cur INTO @perm_name
						END
					CLOSE grant_perm_cur
					DEALLOCATE grant_perm_cur

					OPEN revoke_perm_cur
					FETCH NEXT FROM revoke_perm_cur INTO @perm_name
					WHILE @@FETCH_STATUS = 0
						BEGIN
							SET @stmt = 'REVOKE ' + @perm_name + ' FROM ' + QuoteName(@username)
							EXEC (@stmt)
							FETCH NEXT FROM revoke_perm_cur INTO @perm_name
						END
					CLOSE revoke_perm_cur
					DEALLOCATE revoke_perm_cur
					`
	return c.
	setDatabase(&permissions.DatabaseName).
		ExecContext(ctx, cmd,
			sql.Named("username", permissions.UserName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}

func (c *Connector) DeleteDatabasePermissions(ctx context.Context, permissions *model.DatabasePermissions) error {
	cmd := `DECLARE @stmt nvarchar(max)
					DECLARE perm_cur CURSOR FOR SELECT value FROM String_Split(@permissions, ',')
					DECLARE @permission_name nvarchar(max)
					OPEN perm_cur
					FETCH NEXT FROM perm_cur INTO @permission_name
					WHILE @@FETCH_STATUS = 0
						BEGIN
							SET @stmt = 'REVOKE ' + @permission_name + ' FROM ' + QuoteName(@username)
							EXEC (@stmt)
							FETCH NEXT FROM perm_cur INTO @permission_name
						END
					CLOSE perm_cur
					DEALLOCATE perm_cur
					`
	return c.
	setDatabase(&permissions.DatabaseName).
		ExecContext(ctx, cmd,
			sql.Named("username", permissions.UserName),
			sql.Named("permissions", strings.Join(permissions.Permissions, ",")),
		)
}
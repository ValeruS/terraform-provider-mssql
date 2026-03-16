package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabaseRole(ctx context.Context, database, roleName string) (*model.DatabaseRole, error) {
	cmd := `SELECT
				dp2.principal_id,
				dp2.name,
				dp2.owning_principal_id,
				CASE
					WHEN @@VERSION LIKE 'Microsoft SQL Azure%'
						AND @database = 'master'
						AND (@ownerName = 'dbo' OR @ownerName = '') THEN ''
					ELSE dp1.name
				END AS ownerName
			FROM [sys].[database_principals] dp1
			INNER JOIN [sys].[database_principals] dp2
				ON dp1.principal_id = dp2.owning_principal_id
			WHERE dp2.type = 'R'
				AND dp2.name = @roleName`
	var role model.DatabaseRole
	err := c.
		setDatabase(&database).
		QueryRowContext(ctx, cmd,
			func(r *sql.Row) error {
				return r.Scan(&role.RoleID, &role.RoleName, &role.OwnerId, &role.OwnerName)
			},
			sql.Named("database", database),
			sql.Named("roleName", roleName),
			sql.Named("ownerName", role.OwnerName),
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (c *Connector) CreateDatabaseRole(ctx context.Context, database, roleName string, ownerName string) error {
	cmd := `DECLARE @sql nvarchar(max);
			IF @ownerName = 'dbo' OR @ownerName = ''
				BEGIN
					SET @sql = 'CREATE ROLE ' + QuoteName(@roleName)
				END
			ELSE
				BEGIN
					SET @sql = 'CREATE ROLE ' + QuoteName(@roleName) + ' AUTHORIZATION ' + QuoteName(@ownerName)
				END
			EXEC (@sql);`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
			sql.Named("ownerName", ownerName),
			sql.Named("database", database),
		)
}

func (c *Connector) DeleteDatabaseRole(ctx context.Context, database, roleName string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'IF EXISTS (SELECT 1 FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE [name] = ' + QuoteName(@roleName, '''') + ') ' +
						'DROP ROLE ' + QuoteName(@roleName)
			EXEC (@sql)`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("database", database),
			sql.Named("roleName", roleName),
		)
}

func (c *Connector) UpdateDatabaseRoleName(ctx context.Context, database string, newroleName string, oldroleName string) error {
	cmd := `DECLARE @sql NVARCHAR(max)
			SET @sql = 'ALTER ROLE ' + QuoteName(@oldroleName) + ' WITH NAME = ' + QuoteName(@newroleName)
			EXEC (@sql)`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("database", database),
			sql.Named("newroleName", newroleName),
			sql.Named("oldroleName", oldroleName),
		)
}

func (c *Connector) UpdateDatabaseRoleOwner(ctx context.Context, database string, roleName string, ownerName string) error {
	cmd := `DECLARE @sql NVARCHAR(max)
			IF @ownerName = 'dbo' OR @ownerName = ''
				BEGIN
					SET @ownerName = (SELECT USER_NAME())
				END
			SET @sql = 'ALTER AUTHORIZATION ON ROLE:: [' + @roleName + '] TO [' + @ownerName + ']'
			EXEC (@sql)`

	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("database", database),
			sql.Named("ownerName", ownerName),
			sql.Named("roleName", roleName),
		)
}

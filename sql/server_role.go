package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetServerRole(ctx context.Context, roleName string) (*model.ServerRole, error) {
	cmd := `SELECT dp2.principal_id, dp2.name, dp2.owning_principal_id, dp1.name AS ownerName
			FROM [sys].[server_principals] dp1
			INNER JOIN [sys].[server_principals] dp2 ON dp1.principal_id = dp2.owning_principal_id
			WHERE dp2.type = 'R' AND dp2.name = @roleName`
	var role model.ServerRole
	err := c.
		QueryRowContext(ctx, cmd,
			func(r *sql.Row) error {
				return r.Scan(&role.RoleID, &role.RoleName, &role.OwnerId, &role.OwnerName)
			},
			sql.Named("roleName", roleName),
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (c *Connector) CreateServerRole(ctx context.Context, roleName string, ownerName string) error {
	cmd := `DECLARE @sql nvarchar(max);
			IF @ownerName = ''
				BEGIN
					SET @sql = 'CREATE SERVER ROLE ' + QuoteName(@roleName)
				END
			ELSE
				BEGIN
					SET @sql = 'CREATE SERVER ROLE ' + QuoteName(@roleName) + ' AUTHORIZATION ' + QuoteName(@ownerName)
				END
			EXEC (@sql);`

	return c.
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
			sql.Named("ownerName", ownerName),
		)
}

func (c *Connector) DeleteServerRole(ctx context.Context, roleName string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'IF EXISTS (SELECT 1 FROM [sys].[server_principals] WHERE [name] = ' + QuoteName(@roleName, '''') + ') ' +
						'DROP SERVER ROLE ' + QuoteName(@roleName)
			EXEC (@sql)`

	return c.
		ExecContext(ctx, cmd,
			sql.Named("roleName", roleName),
		)
}

func (c *Connector) UpdateServerRoleName(ctx context.Context, newroleName string, oldroleName string) error {
	cmd := `DECLARE @sql NVARCHAR(max)
			SET @sql = 'ALTER SERVER ROLE ' + QuoteName(@oldroleName) + ' WITH NAME = ' + QuoteName(@newroleName)
			EXEC (@sql)`

	return c.
		ExecContext(ctx, cmd,
			sql.Named("newroleName", newroleName),
			sql.Named("oldroleName", oldroleName),
		)
}

func (c *Connector) UpdateServerRoleOwner(ctx context.Context, roleName string, ownerName string) error {
	cmd := `DECLARE @sql NVARCHAR(max)
			IF @ownerName = ''
				BEGIN
					SET @ownerName = (SELECT SUSER_SNAME())
				END
			SET @sql = 'ALTER AUTHORIZATION ON SERVER ROLE:: [' + @roleName + '] TO [' + @ownerName + ']'
			EXEC (@sql)`

	return c.
		ExecContext(ctx, cmd,
			sql.Named("ownerName", ownerName),
			sql.Named("roleName", roleName),
		)
}

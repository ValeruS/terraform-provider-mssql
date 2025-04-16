package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetEntraIDLogin(ctx context.Context, name string) (*model.EntraIDLogin, error) {
	var login model.EntraIDLogin
	err := c.QueryRowContext(ctx,
		"SELECT name, default_database_name, default_language_name, principal_id, CONVERT(VARCHAR(85), [sid], 1) FROM [master].[sys].[server_principals] WHERE [type] NOT IN ('G', 'R') and [name] = @name",
		func(r *sql.Row) error {
			return r.Scan(&login.LoginName, &login.DefaultDatabase, &login.DefaultLanguage, &login.PrincipalID, &login.Sid)
		},
		sql.Named("name", name),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &login, nil
}

func (c *Connector) CreateEntraIDLogin(ctx context.Context, name, objectId string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			SET @stmt = 'CREATE LOGIN ' + QuoteName(@name) + ' FROM EXTERNAL PROVIDER'
			IF @@VERSION LIKE 'Microsoft SQL Azure%'
				BEGIN
					IF @objectId != ''
						BEGIN
							SET @stmt = @stmt + ' WITH OBJECT_ID = ' + QuoteName(@objectId, '''')
						END
				END
			EXEC (@stmt)`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("name", name),
			sql.Named("objectId", objectId),
		)
}

func (c *Connector) DeleteEntraIDLogin(ctx context.Context, name string) error {
	// Try to kill sessions but continue even if it fails
	_ = c.killSessionsForLogin(ctx, name)

	cmd := `DECLARE @stmt nvarchar(max)
			SET @stmt = 'IF EXISTS (SELECT 1 FROM [master].[sys].[server_principals] WHERE [name] = ' + QuoteName(@name, '''') + ') ' +
						'DROP LOGIN ' + QuoteName(@name)
			EXEC (@stmt)`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("name", name),
		)
}

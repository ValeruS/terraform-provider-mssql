package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabase(ctx context.Context, databaseName string) (*model.Database, error) {
	cmd := `SELECT database_id, name, collation_name, compatibility_level
			FROM [sys].[databases]
			WHERE [name] = @databaseName`
	var db model.Database
	err := c.QueryRowContext(ctx, cmd,
		func(r *sql.Row) error {
			return r.Scan(&db.DatabaseID, &db.DatabaseName, &db.Collation, &db.CompatibilityLevel)
		},
		sql.Named("databaseName", databaseName),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &db, nil
}

func (c *Connector) CreateDatabase(ctx context.Context, databaseName string, collation string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName)
			IF @collation != ''
				BEGIN
					SET @sql = @sql + ' COLLATE ' + @collation
				END
			EXEC (@sql)`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("databaseName", databaseName),
			sql.Named("collation", collation),
		)
}

func (c *Connector) UpdateDatabase(ctx context.Context, databaseName string, newdatabaseName string, collation string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'ALTER DATABASE ' + QuoteName(@databaseName) + ' '
			IF @newdatabaseName != ''
				BEGIN
					SET @sql = @sql + ' MODIFY NAME = ' + QuoteName(@newdatabaseName)
				END
			IF @collation != ''
				BEGIN
					SET @sql = @sql + ' COLLATE ' + @collation
				END
			EXEC (@sql)`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("databaseName", databaseName),
			sql.Named("newdatabaseName", newdatabaseName),
			sql.Named("collation", collation),
		)
}

func (c *Connector) DeleteDatabase(ctx context.Context, databaseName string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'IF EXISTS (SELECT 1 FROM [sys].[databases] WHERE [name] = ' + QuoteName(@databaseName, '''') + ') ' +
						'ALTER DATABASE ' + QuoteName(@databaseName) + ' SET SINGLE_USER WITH ROLLBACK IMMEDIATE; ' +
						'DROP DATABASE ' + QuoteName(@databaseName)
			EXEC (@sql)`
	return c.
		ExecContext(ctx, cmd,
			sql.Named("databaseName", databaseName),
		)
}

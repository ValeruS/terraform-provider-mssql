package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetDatabase(ctx context.Context, databaseName string) (*model.Database, error) {
	cmd := `SELECT database_id, name, collation_name
			FROM [sys].[databases]
			WHERE [name] = @databaseName`
	var db model.Database
	err := c.QueryRowContext(ctx, cmd,
		func(r *sql.Row) error {
			return r.Scan(&db.DatabaseID, &db.DatabaseName, &db.Collation)
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
			IF @collation = ''
				BEGIN
					SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName)
				END
			ELSE
				BEGIN
					SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName) + ' COLLATE ' + @collation
				END
			EXEC (@sql)`
	return c.ExecContext(ctx, cmd,
		sql.Named("databaseName", databaseName),
		sql.Named("collation", collation),
	)
}

func (c *Connector) UpdateDatabaseCollation(ctx context.Context, databaseName string, collation string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'ALTER DATABASE ' + QuoteName(@databaseName) + ' COLLATE ' + @collation
			EXEC (@sql)`
	return c.ExecContext(ctx, cmd,
		sql.Named("databaseName", databaseName),
		sql.Named("collation", collation),
	)
}

func (c *Connector) DeleteDatabase(ctx context.Context, databaseName string) error {
	cmd := `DECLARE @sql nvarchar(max)
			SET @sql = 'IF EXISTS (SELECT 1 FROM [sys].[databases] WHERE [name] = ' + QuoteName(@databaseName, '''') + ') ' +
						'DROP DATABASE ' + QuoteName(@databaseName)
			EXEC (@sql)`
	return c.ExecContext(ctx, cmd,
		sql.Named("databaseName", databaseName),
	)
}

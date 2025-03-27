package sql

import (
	"context"
	"database/sql"

	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
)

func (c *Connector) GetAzureExternalDatasource(ctx context.Context, database, datasourcename string) (*model.AzureExternalDatasource, error) {
	var extds model.AzureExternalDatasource
	var rdbname sql.NullString
	err := c.
		setDatabase(&database).
		QueryRowContext(ctx,
		"SELECT eds.name, eds.data_source_id, eds.location, eds.type_desc, dsc.name, eds.credential_id, eds.database_name FROM [sys].[external_data_sources] eds INNER JOIN [sys].[database_scoped_credentials] dsc ON dsc.credential_id = eds.credential_id AND eds.name = @datasourcename",
		func(r *sql.Row) error {
			return r.Scan(&extds.DataSourceName, &extds.DataSourceId, &extds.Location, &extds.TypeStr, &extds.CredentialName, &extds.CredentialId, &rdbname)
		},
		sql.Named("datasourcename", datasourcename),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	extds.RDatabaseName = rdbname.String
	return &extds, nil
}

func (c *Connector) CreateAzureExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, typestr, rdatabasename string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			SET @stmt = 'CREATE EXTERNAL DATA SOURCE ' + QuoteName(@datasourcename) + ' WITH (LOCATION = ' + QuoteName(@location, '''') + ', CREDENTIAL = ' + QuoteName(@credentialname) + ', TYPE = ' + @typestr
			IF @rdatabasename != ''
				BEGIN
					SET @stmt = @stmt + ', DATABASE_NAME = ' + QuoteName(@rdatabasename, '''')
				END
			SET @stmt = @stmt + ')'
			EXEC (@stmt)`
	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("datasourcename", datasourcename),
			sql.Named("location", location),
			sql.Named("credentialname", credentialname),
			sql.Named("typestr", typestr),
			sql.Named("rdatabasename", rdatabasename),
		)
}

func (c *Connector) UpdateAzureExternalDatasource(ctx context.Context, database, datasourcename, location, credentialname, rdatabasename string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			SET @stmt = 'ALTER EXTERNAL DATA SOURCE ' + QuoteName(@datasourcename) + ' SET LOCATION = ' + QuoteName(@location, '''') + ', CREDENTIAL = ' + QuoteName(@credentialname)
			IF @rdatabasename != ''
				BEGIN
					SET @stmt = @stmt + ', DATABASE_NAME = ' + QuoteName(@rdatabasename, '''')
				END
			EXEC (@stmt)`
	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("datasourcename", datasourcename),
			sql.Named("location", location),
			sql.Named("credentialname", credentialname),
			sql.Named("rdatabasename", rdatabasename),
		)
}

func (c *Connector) DeleteAzureExternalDatasource(ctx context.Context, database, datasourcename string) error {
	cmd := `DECLARE @stmt nvarchar(max)
			SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[external_data_sources] WHERE [name] = ' + QuoteName(@datasourcename, '''') + ') ' +
						'DROP EXTERNAL DATA SOURCE ' + QuoteName(@datasourcename)
			EXEC (@stmt)`
	return c.
		setDatabase(&database).
		ExecContext(ctx, cmd,
			sql.Named("datasourcename", datasourcename),
		)
}

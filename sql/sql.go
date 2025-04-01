package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/ValeruS/terraform-provider-mssql/mssql/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/pkg/errors"
)

type factory struct{}

func GetFactory() model.ConnectorFactory {
	return new(factory)
}

func (f factory) GetConnector(prefix string, data *schema.ResourceData) (interface{}, error) {
	if len(prefix) > 0 {
		prefix = prefix + ".0."
	}

	connector := &Connector{
		Host:    data.Get(prefix + "host").(string),
		Port:    data.Get(prefix + "port").(string),
		Timeout: data.Timeout(schema.TimeoutRead),
	}

	if admin, ok := data.GetOk(prefix + "login.0"); ok {
		admin := admin.(map[string]interface{})
		connector.Login = &LoginUser{
			Username: admin["username"].(string),
			Password: admin["password"].(string),
		}
	}

	if admin, ok := data.GetOk(prefix + "azure_login.0"); ok {
		admin := admin.(map[string]interface{})
		connector.AzureLogin = &AzureLogin{
			TenantID:     admin["tenant_id"].(string),
			ClientID:     admin["client_id"].(string),
			ClientSecret: admin["client_secret"].(string),
		}
	}

	if admin, ok := data.GetOk(prefix + "azuread_managed_identity_auth.0"); ok {
		admin := admin.(map[string]interface{})
		connector.FedauthMSI = &FedauthMSI{
			UserID: admin["user_id"].(string),
		}
	}

	return connector, nil
}

type Connector struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Database   string `json:"database"`
	Login      *LoginUser
	AzureLogin *AzureLogin
	FedauthMSI *FedauthMSI
	Timeout    time.Duration `json:"timeout,omitempty"`
	Token      string
}

type LoginUser struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AzureLogin struct {
	TenantID     string `json:"tenant_id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type FedauthMSI struct {
	UserID string `json:"user_id,omitempty"`
}

func (c *Connector) PingContext(ctx context.Context) error {
	db, err := c.db()
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "In ping")
	}

	return nil
}

// Execute an SQL statement and ignore the results
func (c *Connector) ExecContext(ctx context.Context, command string, args ...interface{}) error {
	db, err := c.db()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, command, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) QueryContext(ctx context.Context, query string, scanner func(*sql.Rows) error, args ...interface{}) error {
	db, err := c.db()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = scanner(rows)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) QueryRowContext(ctx context.Context, query string, scanner func(*sql.Row) error, args ...interface{}) error {
	db, err := c.db()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		return row.Err()
	}

	return scanner(row)
}

func (c *Connector) db() (*sql.DB, error) {
	if c == nil {
		panic("No connector")
	}
	conn, err := c.connector()
	if err != nil {
		return nil, err
	}
	if db, err := connectLoop(conn, c.Timeout); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

func (c *Connector) connector() (driver.Connector, error) {
	query := url.Values{}
	host := fmt.Sprintf("%s:%s", c.Host, c.Port)
	if c.Database != "" {
		query.Set("database", c.Database)
	}
	if c.Login != nil || c.AzureLogin != nil {
		connectionString := (&url.URL{
			Scheme:   "sqlserver",
			User:     c.userPassword(),
			Host:     host,
			RawQuery: query.Encode(),
		}).String()
		if c.Login != nil {
			return mssql.NewConnector(connectionString)
		}
		return mssql.NewAccessTokenConnector(connectionString, func() (string, error) { return c.tokenProvider() })
	}
	if c.FedauthMSI != nil {
		query.Set("fedauth", "ActiveDirectoryManagedIdentity")
		if c.FedauthMSI.UserID != "" {
			query.Set("user id", c.FedauthMSI.UserID)
		}
	} else {
		query.Set("fedauth", "ActiveDirectoryDefault")
	}
	connectionString := (&url.URL{
		Scheme:   "sqlserver",
		Host:     host,
		RawQuery: query.Encode(),
	}).String()
	return azuread.NewConnector(connectionString)
}

func (c *Connector) userPassword() *url.Userinfo {
	if c.Login != nil {
		return url.UserPassword(c.Login.Username, c.Login.Password)
	}
	return nil
}

func (c *Connector) tokenProvider() (string, error) {
	const resourceID = "https://database.windows.net/"

	admin := c.AzureLogin
	cred, err := azidentity.NewClientSecretCredential(
		admin.TenantID,
		admin.ClientID,
		admin.ClientSecret,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{resourceID + "/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	c.Token = token.Token

	return token.Token, nil
}

func connectLoop(connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("db connection failed after %s timeout", timeout)

		case <-ticker.C:
			db, err := connect(connector)
			if err == nil {
				return db, nil
			}
			if strings.Contains(strings.ToLower(err.Error()), "login failed") {
				return nil, err
			}
			if strings.Contains(strings.ToLower(err.Error()), "login error") {
				return nil, err
			}
			if strings.Contains(err.Error(), "error retrieving access token") {
				return nil, err
			}
			if strings.Contains(err.Error(), "AuthenticationFailedError") {
				return nil, err
			}
			if strings.Contains(err.Error(), "credential") {
				return nil, err
			}
			if strings.Contains(err.Error(), "request failed") {
				return nil, err
			}
			log.Println(errors.Wrap(err, "failed to connect to database"))
		}
	}
}

func connect(connector driver.Connector) (*sql.DB, error) {
	db := sql.OpenDB(connector)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (c *Connector) setDatabase(database *string) *Connector {
	if *database == "" {
		*database = "master"
	}
	c.Database = *database
	return c
}

func (c *Connector) GetMSSQLVersion(ctx context.Context) (string, error) {
	var version string
	err := c.QueryRowContext(ctx,"SELECT @@VERSION", func(r *sql.Row) error {
		return r.Scan(&version)
	},)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return version, nil
}

// DatabaseExists checks if a database exists in SQL Server
func (c *Connector) DatabaseExists(ctx context.Context, database string) (bool, error) {
	cmd := `
		SELECT COUNT(1) 
		FROM sys.databases 
		WHERE name = @p1
	`
	var count int
	err := c.QueryRowContext(ctx, cmd, func(r *sql.Row) error {
		return r.Scan(&count)
	}, sql.Named("p1", database))
	if err != nil {
		return false, errors.Wrapf(err, "error checking if database [%s] exists", database)
	}
	
	return count > 0, nil
}
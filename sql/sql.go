package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
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

	if chainAuthList, ok := data.Get(prefix + "azuread_default_chain_auth").([]interface{}); ok && len(chainAuthList) > 0 && chainAuthList[0] != nil {
		chainAuth := chainAuthList[0].(map[string]interface{})
		if useOidc, _ := chainAuth["use_oidc"].(bool); useOidc {
			connector.FedauthOIDC = &FedauthOIDC{
				TenantID:          os.Getenv("ARM_TENANT_ID"),
				ClientID:          os.Getenv("ARM_CLIENT_ID"),
				OIDCToken:         os.Getenv("ARM_OIDC_TOKEN"),
				OIDCTokenFilePath: os.Getenv("ARM_OIDC_TOKEN_FILE_PATH"),
			}
		} else if credentialType, _ := chainAuth["credential_type"].(string); credentialType != "" {
			connector.FedauthCredential = credentialTypes[credentialType]
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

// credentialTypes maps azuread_default_chain_auth.credential_type to the go-mssqldb fedauth value.
var credentialTypes = map[string]string{
	"az_cli":              azuread.ActiveDirectoryAzCli,
	"azure_developer_cli": azuread.ActiveDirectoryAzureDeveloperCli,
	"azure_pipelines":     azuread.ActiveDirectoryAzurePipelines,
	"environment":         azuread.ActiveDirectoryEnvironment,
	"workload_identity":   azuread.ActiveDirectoryWorkloadIdentity,
}

type Connector struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Database    string `json:"database"`
	Login       *LoginUser
	AzureLogin  *AzureLogin
	FedauthOIDC *FedauthOIDC
	FedauthMSI  *FedauthMSI
	// FedauthCredential is the fedauth value used instead of ActiveDirectoryDefault, if set.
	FedauthCredential string
	Timeout           time.Duration `json:"timeout,omitempty"`
	Token             string
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

type FedauthOIDC struct {
	TenantID          string `json:"tenant_id,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	OIDCToken         string `json:"oidc_token,omitempty"`
	OIDCTokenFilePath string `json:"oidc_token_file_path,omitempty"`
}

type FedauthMSI struct {
	UserID string `json:"user_id,omitempty"`
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
	if c.FedauthOIDC != nil {
		connectionString := (&url.URL{
			Scheme:   "sqlserver",
			Host:     host,
			RawQuery: query.Encode(),
		}).String()
		return mssql.NewAccessTokenConnector(connectionString, func() (string, error) { return c.oidcTokenProvider() })
	}
	if c.FedauthMSI != nil {
		query.Set("fedauth", "ActiveDirectoryManagedIdentity")
		if c.FedauthMSI.UserID != "" {
			query.Set("user id", c.FedauthMSI.UserID)
		}
	} else if c.FedauthCredential != "" {
		query.Set("fedauth", c.FedauthCredential)
		if c.FedauthCredential == azuread.ActiveDirectoryAzurePipelines {
			setAzurePipelinesParams(query)
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

// setAzurePipelinesParams passes the variables set by TerraformTask@5 (workload identity
// federation with ID token refresh) to go-mssqldb. When they are not all set, the driver
// falls back to AZURESUBSCRIPTION_* and SYSTEM_ACCESSTOKEN, as set by AzureCLI@2.
func setAzurePipelinesParams(query url.Values) {
	clientID := os.Getenv("ARM_CLIENT_ID")
	serviceConnectionID := os.Getenv("ARM_OIDC_AZURE_SERVICE_CONNECTION_ID")
	systemToken := os.Getenv("ARM_OIDC_REQUEST_TOKEN")
	if clientID == "" || serviceConnectionID == "" || systemToken == "" {
		return
	}

	userID := clientID
	if tenantID := os.Getenv("ARM_TENANT_ID"); tenantID != "" {
		userID += "@" + tenantID
	}
	query.Set("user id", userID)
	query.Set("serviceconnectionid", serviceConnectionID)
	query.Set("systemtoken", systemToken)
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

// oidcGetAssertion resolves the OIDC token from either the inline value or a file.
func (c *Connector) oidcGetAssertion(ctx context.Context) (string, error) {
	oidc := c.FedauthOIDC
	if oidc.OIDCToken != "" {
		return oidc.OIDCToken, nil
	}
	if oidc.OIDCTokenFilePath != "" {
		data, err := os.ReadFile(oidc.OIDCTokenFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read oidc_token_file_path: %v", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", fmt.Errorf("azuread_default_chain_auth with use_oidc=true requires ARM_OIDC_TOKEN or ARM_OIDC_TOKEN_FILE_PATH")
}

func (c *Connector) oidcTokenProvider() (string, error) {
	const resourceID = "https://database.windows.net/"

	oidc := c.FedauthOIDC
	cred, err := azidentity.NewClientAssertionCredential(oidc.TenantID, oidc.ClientID, c.oidcGetAssertion, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create federated credential: %v", err)
	}

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{resourceID + "/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	return token.Token, nil
}

// transientErrors are SQL Server error numbers worth retrying while the server or database starts.
var transientErrors = map[int32]bool{
	18401: true, // login failed: server is in script upgrade mode (container just started)
	40197: true, // Azure SQL: service error processing the request
	40501: true, // Azure SQL: service is busy
	40613: true, // Azure SQL: database not currently available (e.g. serverless resume)
	49918: true, // Azure SQL: not enough resources to process the request
	49919: true, // Azure SQL: too many create or update operations in progress
	49920: true, // Azure SQL: too many operations in progress
}

// connectLoop retries while the server is unreachable or starting. Any other error, such as a
// missing credential or a failed login, is returned immediately.
func connectLoop(connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	lastErr := errors.New("no connection attempt made")
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("%w (retried for %s)", lastErr, timeout)

		case <-ticker.C:
			db, err := connect(connector)
			if err == nil {
				return db, nil
			}
			if !isTransient(err) {
				return nil, err
			}
			lastErr = err
			log.Println(errors.Wrap(err, "failed to connect to database, retrying"))
		}
	}
}

func isTransient(err error) bool {
	var netErr net.Error
	var streamErr mssql.StreamError
	var sqlErr mssql.Error
	switch {
	case errors.As(err, &netErr), errors.As(err, &streamErr),
		errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return true
	case errors.As(err, &sqlErr):
		return transientErrors[sqlErr.Number]
	}
	return false
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
	err := c.QueryRowContext(ctx, "SELECT @@VERSION", func(r *sql.Row) error {
		return r.Scan(&version)
	})
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
	scan := func(r *sql.Row) error {
		return r.Scan(&count)
	}
	err := c.QueryRowContext(ctx, cmd, scan, sql.Named("p1", database))
	var sqlErr mssql.Error
	if err != nil && errors.As(err, &sqlErr) && sqlErr.Number == 18456 && c.Database != database {
		// No access to master, e.g. an Azure AD identity with a user only in the target
		// database. Connecting to the database itself shows that it exists.
		target := *c
		target.Database = database
		err = target.QueryRowContext(ctx, cmd, scan, sql.Named("p1", database))
	}
	if err != nil {
		return false, errors.Wrapf(err, "error checking if database [%s] exists", database)
	}

	return count > 0, nil
}

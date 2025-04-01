package model

type AzureExternalDatasource struct {
	DatabaseName   string
	DataSourceName string
	DataSourceId   int
	Location       string
	TypeStr        string
	CredentialName string
	CredentialId   int
	RDatabaseName  string
}

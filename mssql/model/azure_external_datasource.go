package model

type AzureExternalDatasource struct {
	DatabaseName   string
	DataSourceName string
	DataSourceId   int
	Location       string
	TypeDesc       string
	CredentialName string
	CredentialId   int
	RDatabaseName  string
}
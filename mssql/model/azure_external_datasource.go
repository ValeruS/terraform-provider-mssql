package model

type AzureExternalDatasource struct {
	DataSourceName string
	DataSourceId   int
	Location       string
	TypeStr        string
	CredentialName string
	CredentialId   int
	RDatabaseName  string
}

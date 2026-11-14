package model

type DatabaseSchemaPermission struct {
	DatabaseName  string
	SchemaName    string
	PrincipalName string
	PrincipalID   int
	SchemaID      int
	Permissions   []string
}

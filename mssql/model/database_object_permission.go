package model

type DatabaseObjectPermission struct {
	DatabaseName  string
	SchemaName    string
	ObjectName    string
	ColumnName    string
	PrincipalName string
	PrincipalID   int
	ObjectID      int
	Permissions   []string
}

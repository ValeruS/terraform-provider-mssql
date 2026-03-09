package model

// ServerRole represents a SQL Server role
type ServerRole struct {
	RoleID    int
	RoleName  string
	OwnerName string
	OwnerId   int
}

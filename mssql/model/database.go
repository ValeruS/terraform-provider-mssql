package model

// Database represents a SQL Server database
type Database struct {
	DatabaseID   int
	DatabaseName string
	Collation    string
}

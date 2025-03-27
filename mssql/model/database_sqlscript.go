package model

// DatabaseSQLScript represents a SQL Server database SQL script
type DatabaseSQLScript struct {
	Script            string
	ScriptFile        string
	VerifyObject      string
	LastExecutionHash string
}

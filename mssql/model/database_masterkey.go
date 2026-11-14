package model

type DatabaseMasterkey struct {
	KeyName        string
	KeyGuid        string
	SymmetricKeyID int
	KeyLength      int
	KeyAlgorithm   string
	AlgorithmDesc  string
	PrincipalID    int
}

package bootstrap

type Config interface {
	GetAddress() *string
	GetStoreInterval() *int
	GetFileStoragePath() *string
	GetRestore() *bool
	GetDatabaseDSN() *string
	GetMigrationsPath() *string
	GetCountRetrySave() *int
	GetSecretKey() *string
	GetAuditFilePath() *string
	GetAuditServiceURL() *string
	GetCryptoKey() *string
	GetEnableHTTPS() *bool
	GetTLSCertFile() *string
	GetTLSKeyFile() *string
	GetTrustedSubnet() *string
}

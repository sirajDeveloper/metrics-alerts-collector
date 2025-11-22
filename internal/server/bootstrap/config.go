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
}

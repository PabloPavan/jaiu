package config

type Config struct {
	Redis       RedisConfig
	R2          R2Config
	StorageType StorageType
	LocalDir    string
	QueueName   string
	OriginalKey string
	Sizes       []ImageSize
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type R2Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

type ImageSize struct {
	Name   string
	Width  int
	Height int
}

// StorageType selects the backing store for images.
type StorageType string

const (
	StorageR2    StorageType = "r2"
	StorageLocal StorageType = "local"
)

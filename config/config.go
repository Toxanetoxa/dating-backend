package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config application configuration
	Config struct {
		App
		Log
		HTTP
		Auth
		PG
		Redis
		S3
		Smtp
		Firebase
		YouKassa
		VK
	}

	// App application info
	App struct {
		Name    string `env-required:"true" env:"APP_NAME"`
		Version string `env-required:"true" env:"APP_VERSION"`
	}

	// HTTP transport config
	HTTP struct {
		Port int `env-required:"true" env:"HTTP_PORT" env-default:"8080"`
	}

	// Log logger config
	Log struct {
		Level    string `env-required:"true" env:"LOG_LEVEL" env-default:"debug"`
		BodyDump bool   `env:"ENABLE_BODY_DUMP" env-default:"false"`
	}

	// PG postgres connection config
	PG struct {
		PoolMax int `env-required:"true" env:"PG_POOL_MAX"`
		// URL      string `env-required:"true" env:"PG_URL"`
		Host     string `env-required:"true" env:"PG_HOST"`
		Port     int    `env-required:"true" env:"PG_PORT"`
		User     string `env-required:"true" env:"PG_USER"`
		Password string `env-required:"true" env:"PG_PASSWORD"`
		Db       string `env-required:"true" env:"PG_DB"`
		SSL      string `env-required:"true" env:"PG_SSL"`
	}

	// Redis connection config
	Redis struct {
		Addr     string `env-required:"true" env:"RDB_ADDR"`
		Password string `env-required:"true" env:"RDB_PASSWORD"`
		DB       int    `env-required:"true" env:"RDB_DB"`
	}

	S3 struct {
		Endpoint     string `env-required:"true" env:"S3_ENDPOINT"`
		Secure       bool   `env-required:"true" env:"S3_SECURE" env-default:"true"`
		AccessKey    string `env-required:"true" env:"S3_ACCESS_KEY"`
		SecretKey    string `env-required:"true" env:"S3_SECRET_KEY"`
		BucketName   string `env-required:"true" env:"S3_BUCKET_NAME"`
		BaseFilesURL string `env-required:"true" env:"S3_FILES_BASE_URL"`
	}

	Auth struct {
		TokenTTL    time.Duration `env:"TOKEN_TTL" env-default:"1h"`
		AuthCodeTTL time.Duration `env:"AUTH_CODE_TTL" env-default:"5m"`
	}

	Smtp struct {
		Enable   bool   `env:"SMTP_ENABLE" env-default:"false"`
		Host     string `env:"SMTP_HOST"`
		Port     int    `env:"SMTP_PORT"`
		User     string `env:"SMTP_USER"`
		Password string `env:"SMTP_PASSWORD"`
		From     string `env:"SMTP_FROM"`
	}

	Firebase struct {
		Enable              bool   `env:"FIREBASE_ENABLE" env-default:"false"`
		JsonCredentialsPath string `env:"FIREBASE_JSON_CREDENTIALS_PATH"`
	}

	YouKassa struct {
		BaseURL     string `env:"YOUKASSA_API_URL" env-required:"true"`
		ShopID      string `env:"YOUKASSA_SHOP_ID" env-required:"true"`
		SecretKey   string `env:"YOUKASSA_SECRET_KEY" env-required:"true"`
		RedirectUrl string `env:"YOUKASSA_REDIRECT_URL" env-required:"true"`
		CallbackUrl string `env:"YOUKASSA_CALLBACK_URL" env-required:"true"`
	}

	VK struct {
		BaseURL string `env:"VK_API_URL" env-required:"true"`
		AppId   string `env:"VK_APP_ID" env-required:"true"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

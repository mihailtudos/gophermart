package config

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env"
)

const (
	defaultLoggerlevel = "info"

	defaultHTTPPort           = ":8080"
	defaultHTTPMaxHeaderBytes = 1
	defaultHTTPReadTimeout    = "10s"
	defaultHTTPWriteTimeout   = "10s"

	defaultJWTAccessTokenTTL  = "2h"
	defaultJWTRefreshTokenTTL = "720h"

	defaultAccrualSysAddress = "http://localhost:8000"
)

type (
	LoggerConfig struct {
		Level string `mapstructure:"level"`
	}

	AuthConfig struct {
		JWT          JWTConfig
		PasswordSalt string
	}
	JWTConfig struct {
		AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
		RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
		SigningKey      string
	}
	HTTPConfig struct {
		Port           string        `mapstructure:"port" env:"RUN_ADDRESS"`
		MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
		ReadTimeout    time.Duration `mapstructure:"readTimeout"`
		WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	}

	DBConfig struct {
		DSN string `mapstructure:"dsn" env:"DATABASE_URI"`
	}

	AccrualConfig struct {
		Address string `mapstructure:"address" env:"ACCRUAL_SYSTEM_ADDRESS"`
	}

	config struct {
		Logger  LoggerConfig
		HTTP    HTTPConfig
		DB      DBConfig
		Auth    AuthConfig
		Accrual AccrualConfig
	}
)

var (
	once     sync.Once
	instance *config
)

func NewConfig() *config {
	if instance != nil {
		return instance
	}

	once.Do(func() {
		var cfg config
		setDefaults(&cfg)

		flag.StringVar(&cfg.DB.DSN, "d", cfg.DB.DSN, "Database DSN")
		flag.StringVar(&cfg.HTTP.Port, "a", cfg.HTTP.Port, "HTTP server address")
		flag.StringVar(&cfg.Accrual.Address, "r", cfg.Accrual.Address, "Accrual system address")
		flag.Parse()

		if err := env.Parse(&cfg); err != nil {
			panic(fmt.Errorf("failed to load environment variables: %w", err))
		}

		if envPort := os.Getenv("RUN_ADDRESS"); envPort != "" {
			cfg.HTTP.Port = envPort
		}

		if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
			cfg.DB.DSN = envDB
		}

		if envAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrual != "" {
			cfg.Accrual.Address = envAccrual
		}

		instance = &cfg
	})

	return instance
}

func setDefaults(cfg *config) {
	// logger related defaults
	cfg.Logger.Level = defaultLoggerlevel

	// http server related defaults
	cfg.HTTP.Port = defaultHTTPPort
	cfg.HTTP.MaxHeaderBytes = defaultHTTPMaxHeaderBytes
	assignValueCfgProp(&cfg.HTTP.ReadTimeout, defaultHTTPReadTimeout)
	assignValueCfgProp(&cfg.HTTP.WriteTimeout, defaultHTTPWriteTimeout)

	// auth related defaults
	assignValueCfgProp(&cfg.Auth.JWT.AccessTokenTTL, defaultJWTAccessTokenTTL)
	assignValueCfgProp(&cfg.Auth.JWT.RefreshTokenTTL, defaultJWTRefreshTokenTTL)

	// accrual sys defaults
	cfg.Accrual.Address = defaultAccrualSysAddress
}

func assignValueCfgProp(destination *time.Duration, defaultValue string) {
	duration, err := time.ParseDuration(defaultValue)
	if err != nil {
		panic("failed to set default config value")
	}

	*destination = duration
}

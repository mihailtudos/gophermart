package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/spf13/viper"
)

const (
	defaultLoggerlevel       = "info"
	defaultLoggerDestination = ""
	defaultLoggerChannel     = "stack"

	defaultHTTPPort           = ":8080"
	defaultHTTPMaxHeaderBytes = 1
	defaultHTTPReadTimeout    = "10s"
	defaultHTTPWriteTimeout   = "10s"

	defaultJWTAccessTokenTTL  = "2h"
	defaultJWTRefreshTokenTTL = "720h"

	defaultAccrualSysAddress = "http://localhost:8000"
)

type Config struct {
	Logger  LoggerConfig
	HTTP    HTTPConfig
	DB      *DBConfig
	Auth    AuthConfig
	Accrual AccrualConfig
	ToClose map[string]io.WriteCloser
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Destination string `mapstructure:"destination"`
	Channel     string `mapstructure:"channel"`
}

type AuthConfig struct {
	JWT                    JWTConfig
	PasswordSalt           string
	VerificationCodeLength int `mapstructure:"verificationCodeLength"`
}

type JWTConfig struct {
	AccessTokenTTL  time.Duration `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL time.Duration `mapstructure:"refreshTokenTTL"`
	SigningKey      string
}

type HTTPConfig struct {
	Port           string        `mapstructure:"port" env:"RUN_ADDRESS"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
}

type DBConfig struct {
	DSN string `mapstructure:"dsn" env:"DATABASE_URI"`
}

type AccrualConfig struct {
	Address string `mapstructure:"address" env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig(cfgPath string) (Config, error) {
	// setting up viper with the default config options
	setDefaults()

	// loading the config based on the APP_ENV
	if err := loadConfig(cfgPath); err != nil {
		return Config{}, err
	}

	cfg := Config{
		ToClose: make(map[string]io.WriteCloser),
	}

	if err := unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	// overwrites the static config options with dynamically passed env vars
	if err := overwriteStaticConfig(&cfg); err != nil {
		return Config{}, err
	}

	logFile, err := getLoggerFile(cfg.Logger.Destination, cfg.Logger.Channel)
	if err != nil {
		if errors.Is(err, logger.ErrDestinationNotFound) {
			logFile = os.Stdout
		} else {
			return Config{}, nil
		}
	}

	logger.Init(logFile, cfg.Logger.Level)

	// pushing used files so they can be closed on app shut down
	cfg.ToClose["logger"] = logFile
	return cfg, nil
}

func getLoggerFile(destination, channel string) (*os.File, error) {
	if destination == "" {
		return nil, logger.ErrDestinationNotFound
	}

	loggerFileNmae := logger.DefaultLoggerFileName

	if channel == logger.DailyChanne {
		loggerFileNmae = time.Now().Format("2006-01-02")
	}

	return os.OpenFile(path.Join(destination, loggerFileNmae+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("logger", &cfg.Logger); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("http", &cfg.HTTP); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("db", &cfg.DB); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth", &cfg.Auth.JWT); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("accrual", &cfg.Accrual); err != nil {
		return err
	}

	return nil
}

func loadConfig(cfgPath string) error {
	viper.SetConfigName("main")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgPath)
	return viper.ReadInConfig()
}

func overwriteStaticConfig(cfg *Config) error {
	cfg.Auth.PasswordSalt = os.Getenv("PASSWORD_SALT")
	cfg.Auth.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")

	flag.StringVar(&cfg.DB.DSN, "d", cfg.DB.DSN, "Database DSN")
	flag.StringVar(&cfg.HTTP.Port, "a", cfg.HTTP.Port, "HTTP server address")
	flag.StringVar(&cfg.Accrual.Address, "r", cfg.Accrual.Address, "Accrual system address")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
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

	return nil
}

func setDefaults() {
	// logger related defaults
	viper.SetDefault("logger.level", defaultLoggerlevel)
	viper.SetDefault("logger.destination", defaultLoggerDestination)
	viper.SetDefault("logger.channel", defaultLoggerChannel)

	// http server related defaults
	viper.SetDefault("http.port", defaultHTTPPort)
	viper.SetDefault("http.maxHeaderBytes", defaultHTTPMaxHeaderBytes)
	viper.SetDefault("http.readTimeout", defaultHTTPReadTimeout)
	viper.SetDefault("http.writeTimeout", defaultHTTPWriteTimeout)

	// auth related defaults
	viper.SetDefault("auth.accessTokenTTL", defaultJWTAccessTokenTTL)
	viper.SetDefault("auth.refreshTokenTTL", defaultJWTRefreshTokenTTL)

	// accrual sys defaults
	viper.SetDefault("accrual.address", defaultAccrualSysAddress)
}

func init() {
	godotenv.Load(".env")
}

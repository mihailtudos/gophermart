package config

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConfig
	Http    HttpConfig
	DB      *DBConfig
	Auth    AuthConfig
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

type HttpConfig struct {
	Port           string        `mapstructure:"port"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name"`
	SSLMode  string `mapstructure:"mode"`
}

func NewConfig(cfgPath string) (Config, error) {
	if err := loadConfig(cfgPath); err != nil {
		return Config{}, err
	}

	cfg := Config{
		ToClose: make(map[string]io.WriteCloser),
	}

	if err := unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	margeInEnvConfig(&cfg)

	logFile, err := getLoggerFile(cfg.Logger.Destination, cfg.Logger.Channel)
	if err != nil {
		return Config{}, nil
	}

	logger.Init(logFile, cfg.Logger.Level)

	// pushing usef files so they can be closed on app shut down
	cfg.ToClose["logger"] = logFile
	return cfg, nil
}

func getLoggerFile(destination, channel string) (*os.File, error) {
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

	if err := viper.UnmarshalKey("http", &cfg.Http); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("db", &cfg.DB); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("auth", &cfg.Auth.JWT); err != nil {
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

func margeInEnvConfig(cfg *Config) {
	cfg.Auth.PasswordSalt = os.Getenv("PASSWORD_SALT")
	cfg.Auth.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")
}

func init() {
	godotenv.Load(".env")
}

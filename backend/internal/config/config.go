package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"

	_ "github.com/joho/godotenv/autoload"
)

// parent config struct

type Config struct {
	Primary  Primary        `koanf:"primary" validate:"required"`
	Server   ServerConfig   `koanf:"server" validate:"required"`
	Database DatabaseConfig `koanf:"database" validate:"required"`
	Auth     AuthConfig     `koanf:"auth" validate:"required"`
	Redis    RedisConfig    `koanf:"redis" validate:"required"`
}

// one primary struct for env data
// using struct tags ` ` ,  these majorly used in refelection in go

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

// server configuration

type ServerConfig struct {
	Port               int      `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteReadout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

// database configs

type DatabaseConfig struct {
	Port            int    `koanf:"port" validate:"required"`
	Host            string `koanf:"host" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `konaf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifeTime int    `koanf:"conn_max_life_time" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}

// authentication config  - secret key

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

// redis configuration

type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

// loader config function

func LoadConfig() (*Config, error) {

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	k := koanf.New(".")

	err := k.Load(env.Provider("BOILERPLATE_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "BOILERPLATE_"))
	}), nil)

	if err != nil {
		logger.Fatal().Err(err).Msg("could not load initial env variables")
	}

	// unmarshaling the config , from main config of the app

	mainConfig := &Config{}

	err = k.Unmarshal("", mainConfig)

	if err != nil {
		logger.Fatal().Err(err).Msg("could not unmarshal the main config")
	}

	// next step is to validate the values provided

	validate := validator.New()

	err = validate.Struct(mainConfig)

	if err != nil {
		logger.Fatal().Err(err).Msg("config validation failed")
	}

	return mainConfig, err

}

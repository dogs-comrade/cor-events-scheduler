// internal/config/config.go

package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Address string
	Port    string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	// Значения по умолчанию
	viper.SetDefault("SERVER_ADDRESS", "localhost")
	viper.SetDefault("SERVER_PORT", "8282")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "your_secure_password")
	viper.SetDefault("DB_NAME", "mew")

	config := &Config{
		Server: ServerConfig{
			Address: viper.GetString("SERVER_ADDRESS"),
			Port:    viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
		},
	}

	return config, nil
}

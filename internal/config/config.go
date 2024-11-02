// internal/config/config.go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Analysis AnalysisConfig
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

type AnalysisConfig struct {
	MinTechBreakDuration  int     `mapstructure:"MIN_TECH_BREAK_DURATION"`
	EquipmentComplexity   float64 `mapstructure:"EQUIPMENT_COMPLEXITY_FACTOR"`
	MultidayBuffer        int     `mapstructure:"MULTIDAY_BUFFER_TIME"`
	WeatherRiskMultiplier float64 `mapstructure:"WEATHER_RISK_MULTIPLIER"`
	HumanFactorMultiplier float64 `mapstructure:"HUMAN_FACTOR_MULTIPLIER"`
	EquipmentRiskBase     float64 `mapstructure:"EQUIPMENT_RISK_BASE"`
}

func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	// Server defaults
	viper.SetDefault("SERVER_ADDRESS", "localhost")
	viper.SetDefault("SERVER_PORT", "8282")

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "your_secure_password")
	viper.SetDefault("DB_NAME", "mew")

	// Analysis defaults
	viper.SetDefault("MIN_TECH_BREAK_DURATION", 15)
	viper.SetDefault("EQUIPMENT_COMPLEXITY_FACTOR", 1.5)
	viper.SetDefault("MULTIDAY_BUFFER_TIME", 60)
	viper.SetDefault("WEATHER_RISK_MULTIPLIER", 1.2)
	viper.SetDefault("HUMAN_FACTOR_MULTIPLIER", 1.3)
	viper.SetDefault("EQUIPMENT_RISK_BASE", 0.05)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: Config file not found, using environment variables and defaults: %v\n", err)
	}

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
		Analysis: AnalysisConfig{
			MinTechBreakDuration:  viper.GetInt("MIN_TECH_BREAK_DURATION"),
			EquipmentComplexity:   viper.GetFloat64("EQUIPMENT_COMPLEXITY_FACTOR"),
			MultidayBuffer:        viper.GetInt("MULTIDAY_BUFFER_TIME"),
			WeatherRiskMultiplier: viper.GetFloat64("WEATHER_RISK_MULTIPLIER"),
			HumanFactorMultiplier: viper.GetFloat64("HUMAN_FACTOR_MULTIPLIER"),
			EquipmentRiskBase:     viper.GetFloat64("EQUIPMENT_RISK_BASE"),
		},
	}

	return config, nil
}

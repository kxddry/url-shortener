package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string       `yaml:"env" env-required:"true"`
	Storage    Storage      `yaml:"postgres" env-required:"true"`
	HTTPServer HTTPServer   `yaml:"http_server"`
	Redis      RedisStorage `yaml:"redis" env-required:"true"`
}

type Storage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

type RedisStorage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-default:""`
	Password string `yaml:"password" env-default:""`
	DB       int    `yaml:"db" env-default:"0" validate:"min=0,integer"`
	PoolSize int    `yaml:"pool_size" env-default:"10"`
	Protocol string `yaml:"protocol" env-default:"tcp"`
}

type HTTPServer struct {
	Address     string        `yaml:"host" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	var cfg Config
	// Read the config file
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// validate the config file
	err := cfg.validate()
	if err != nil {
		log.Fatal(err)
	}
	return &cfg
}

func (c *Config) validate() error {
	if c.Env != "local" && c.Env != "dev" && c.Env != "prod" {
		return errors.New("config: invalid env value")
	}
	return nil
}

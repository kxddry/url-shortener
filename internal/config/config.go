package config

import (
	"errors"
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env        string        `yaml:"env" env-required:"true"`
	Storage    Storage       `yaml:"postgres" env-required:"true"`
	HTTPServer HTTPServer    `yaml:"http_server"`
	Redis      RedisStorage  `yaml:"redis" env-required:"true"`
	Clients    ClientsConfig `yaml:"clients"`
	App        App           `yaml:"app" env-required:"true"`
	TokenTTL   time.Duration `yaml:"token_ttl" env-required:"true"`
}

type App struct {
	Name   string `yaml:"name" env-required:"true"`
	Secret string `yaml:"secret" env-required:"true" env:"APP_SECRET"`
	ID     int64
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

type Client struct {
	Address  string        `yaml:"address"`
	Timeout  time.Duration `yaml:"timeout"`
	Retries  int           `yaml:"retries"`
	Insecure bool          `yaml:"insecure"`
}

type ClientsConfig struct {
	SSO Client `yaml:"sso"`
}

type HTTPServer struct {
	Address     string        `yaml:"host" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	return MustLoadByPath(path)
}

func (c *Config) validate() error {
	if c.Env != "local" && c.Env != "dev" && c.Env != "prod" {
		return errors.New("config: invalid env value")
	}
	return nil
}

type Storage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"enable"`
}

type MigrationConfig struct {
	Storage    Storage    `yaml:"storage" env-required:"true"`
	Migrations Migrations `yaml:"migrations" env-required:"true"`
}

type Migrations struct {
	Path string `yaml:"path" env-required:"true"`
}

type GRPCServer struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

func MustLoadMigration() *MigrationConfig {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	return MustLoadMigrationByPath(path)
}

func MustLoadMigrationByPath(path string) *MigrationConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file doesn't exist " + path)
	}
	var res MigrationConfig
	if err := cleanenv.ReadConfig(path, &res); err != nil {
		panic(err)
	}
	return &res
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file doesn't exist " + path)
	}
	var res Config
	if err := cleanenv.ReadConfig(path, &res); err != nil {
		panic(err)
	}
	return &res
}

// get config path from flag or env.
// prioritize flag over env over default
// default: empty string
func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res != "" {
		return res
	}
	env := os.Getenv("CONFIG_PATH")
	return env
}

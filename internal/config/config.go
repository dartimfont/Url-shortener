package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	DatabaseURL string `yaml:"storage_path" env-required:"true" env:"DB_URL"`
	HTTPServer  `yaml:"http_server"`
	Auth        `yaml:"auth"`
}

type Auth struct {
	User     string `yaml:"user" env-required:"true" env:"AUTH_USER"`
	Password string `yaml:"password" env-required:"true" env:"AUTH_PASSWORD"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		log.Print("Using DB_URL from environment")
	}

	return &cfg
}

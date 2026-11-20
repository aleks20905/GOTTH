package config

// config extract envconfig for global
//
import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port              string `envconfig:"SERVERPORT" default:":4000"`
	DatabaseName      string `envconfig:"DATABASE_NAME" default:"goth.db"`
	SessionCookieName string `envconfig:"SESSION_COOKIE_NAME" default:"session"`
	SessionSecret     string `envconfig:"SESSION_SECRET"`
	DatabaseURL       string `envconfig:"DATABASE_URL"`
	StaticDir         string `envconfig:"STATIC_DIR" default:"./static"`
	Agents            string `envconfig:"AGENTS"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

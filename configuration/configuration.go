package configuration

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

// Config is a struct that holds the application configuration
type Config struct {
	Address     string `koanf:"Address"`
	ServiceName string `koanf:"ServiceName"`
	Authority   string `koanf:"Authority"`
	DB          struct {
		Dsn           string `koanf:"Dsn"`
		MaxIdleTimeMS int    `koanf:"MaxIdleTimeMs"`
		MaxOpenConns  int    `koanf:"MaxOpenConns"`
		MaxIdleConns  int    `koanf:"MaxIdleConns"`
	} `koanf:"DB"`
	SMTP struct {
		Host     string `koanf:"Host"`
		Port     int    `koanf:"Port"`
		Username string `koanf:"Username"`
		Password string `koanf:"Password"`
		Sender   string `koanf:"Sender"`
	} `koanf:"SMTP"`
	RabbitMQ struct {
		Host     string `koanf:"Host"`
		Port     int    `koanf:"Port"`
		User     string `koanf:"User"`
		Password string `koanf:"Password"`
	} `koanf:"RabbitMQ"`
}

// LoadConfig reads configuration from a given file and from environment variables
// (i.e. SMTP__Host=...).
func LoadConfig(filePath string) (*Config, error) {
	var config Config

	configReader := koanf.New(".")

	// Load JSON config
	if err := configReader.Load(file.Provider(filePath), json.Parser()); err != nil {
		return nil, err
	}

	// Load environment variables and merge into the loaded config
	configReader.Load(
		env.Provider(
			"",
			"__",
			nil,
		),
		nil,
	)

	err := configReader.Unmarshal("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

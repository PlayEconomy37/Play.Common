package configuration

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Address     string `koanf:"address"`
	ServiceName string `koanf:"serviceName"`
	Authority   string `koanf:"authority"`
	Db          struct {
		Dsn string `koanf:"dsn"`
	} `koanf:"db"`
}

// Reads configuration from file and/or environment variables
func LoadConfig() (Config, error) {
	var config Config

	configReader := koanf.New(".")

	// Load JSON config
	if err := configReader.Load(file.Provider("config/dev.json"), json.Parser()); err != nil {
		return config, err
	}

	// Load environment variables and merge into the loaded config.
	// We lowercase the key, replace `_` with `.` and strip the APP_ prefix.
	configReader.Load(
		env.Provider(
			"APP_",
			".",
			func(s string) string {
				return strings.Replace(
					strings.ToLower(strings.TrimPrefix(s, "APP_")),
					"_",
					".",
					-1,
				)
			},
		),
		nil,
	)

	err := configReader.Unmarshal("", &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

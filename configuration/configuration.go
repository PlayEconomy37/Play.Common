package configuration

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Address     string `koanf:"Address"`
	ServiceName string `koanf:"ServiceName"`
	Db          struct {
		Dsn           string `koanf:"Dsn"`
		MaxIdleTimeMS int    `koanf:"MaxIdleTimeMs"`
		MaxOpenConns  int    `koanf:"MaxOpenConns"`
		MaxIdleConns  int    `koanf:"MaxIdleConns"`
	} `koanf:"Db"`
}

// Reads configuration from file and/or environment variables
func LoadConfig(filePath string) (Config, error) {
	var config Config

	configReader := koanf.New(".")

	// Load JSON config
	if err := configReader.Load(file.Provider(filePath), json.Parser()); err != nil {
		return config, err
	}

	// Load environment variables and merge into the loaded config.
	// We lowercase the key, replace `_` with `.` and strip the APP_ prefix.
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
		return config, err
	}

	return config, nil
}

package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Read the encryption secrets from the path specified in config.json
	secretsPath := viper.GetString("encryption_secrets.path")

	secretsViper := viper.New()
	secretsViper.SetConfigFile(secretsPath)
	secretsViper.SetConfigType("json")

	if err := secretsViper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read encryption secrets: %w", err)
	}

	// Merge secrets into main viper
	if err := viper.MergeConfigMap(secretsViper.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge encryption secrets config: %w", err)
	}

	required := []string{
		"mongodb.username",
		"mongodb.password",
		"mongodb.host",
		"mongodb.port",
		"mongodb.database",
		"mongodb.collection",
		"netbox.url",
		"netbox.token",
		"encryption_secrets.path",
		"enc_key",
		"enc_iv",
	}

	for _, key := range required {
		if !viper.IsSet(key) {
			return fmt.Errorf("missing required config key: %s", key)
		}
	}

	return nil
}

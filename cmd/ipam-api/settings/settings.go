package settings

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
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
	secretsViper.AddConfigPath(".") // Add current directory to search path

	if err := secretsViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return fmt.Errorf("failed to read encryption secrets, file not found : %w", err)
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("failed to read encryption secrets: %w", err)
		}
	}

	// Merge secrets into main viper
	if err := viper.MergeConfigMap(secretsViper.AllSettings()); err != nil {
		return fmt.Errorf("failed to merge encryption secrets config: %w", err)
	}

	viper.Set("mongodb.collection", "addresses") // Set default collection name

	if viper.GetString("mongodb.password_path") != "" {
		secretPath := viper.GetString("mongodb.password_path")
		secret, err := os.ReadFile(secretPath)
		if err != nil {
			return fmt.Errorf("failed to read mongodb password from file: %w", err)
		}
		viper.Set("mongodb.password", string(secret))
	}

	if viper.GetString("netbox.token_path") != "" {
		secretPath := viper.GetString("netbox.token_path")
		secret, err := os.ReadFile(secretPath)
		if err != nil {
			return fmt.Errorf("failed to read Netbox token from file: %w", err)
		}
		viper.Set("netbox.token", string(secret))
	}

	if viper.GetString("splunk.token_path") != "" {
		secretPath := viper.GetString("splunk.token_path")
		secret, err := os.ReadFile(secretPath)
		if err != nil {
			return fmt.Errorf("failed to read Splunk token from file: %w", err)
		}
		viper.Set("splunk.token", string(secret))
	}

	required := []string{
		"mongodb.username",
		"mongodb.password_path",
		"mongodb.host",
		"mongodb.port",
		"mongodb.database",
		"netbox.url",
		"netbox.token_path",
		"encryption_secrets.path",
		"enc_key",
		"enc_iv",
	}

	for _, key := range required {
		if !viper.IsSet(key) {
			return fmt.Errorf("missing required config key: %s", key)
		}
	}

	if viper.GetString("netbox.constraint_tag") != "" {
		constraintTagId, err := netboxservice.GetTagId(viper.GetString("netbox.constraint_tag"))
		if err != nil {
			return fmt.Errorf("failed to get constraint tag ID: %w", err)
		}
		viper.Set("netbox.constraint_tag_id", constraintTagId)
	}

	return nil
}

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

	required := []string{
		"mongodb.username",
		"mongodb.password",
		"mongodb.host",
		"mongodb.port",
		"mongodb.database",
		"mongodb.collection",
		"netbox.url",
		"netbox.token",
		"netbox.prefix_containers.inet_v4",
		"netbox.prefix_containers.inet_v6",
		"netbox.prefix_containers.hnet_private_v4",
		"netbox.prefix_containers.hnet_public_v4",
	}

	for _, key := range required {
		if !viper.IsSet(key) {
			return fmt.Errorf("missing required config key: %s", key)
		}
	}

	return nil
}

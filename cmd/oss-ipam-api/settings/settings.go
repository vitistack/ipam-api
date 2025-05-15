package settings

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	viper.GetString("mongodb.host")

	fmt.Println("MongoDB Host:", viper.GetString("mongodb.host"))
}

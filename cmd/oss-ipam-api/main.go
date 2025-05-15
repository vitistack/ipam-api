package main

import (
	"github.com/NorskHelsenett/oss-ipam-api/cmd/oss-ipam-api/settings"
	"github.com/NorskHelsenett/oss-ipam-api/internal/webserver"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/spf13/viper"
)

func main() {
	// read config.json file
	settings.InitConfig()

	// Initialize MongoDB client
	mongoConfig := mongodb.MongoConfig{
		Host:     viper.GetString("mongodb.host"),
		Username: viper.GetString("mongodb.username"),
		Password: viper.GetString("mongodb.password")}

	mongodb.InitClient(mongoConfig)

	//initialize web server
	webserver.InitHttpServer()
}

// CORTEX ERROR EXAMPLE
// func main() {

// 	// viper.SetConfigFile("config.json")
// 	viper.SetConfigName("config")
// 	viper.SetConfigType("json")
// 	viper.AddConfigPath(".")

// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("OK: Read config")

// 	username := viper.GetString("mongodb.username")
// 	fmt.Println("Username:", username)
// }

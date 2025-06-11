package main

import (
	"log"

	"github.com/vitistack/ipam-api/cmd/ipam-api/settings"
)

func main() {

	err := settings.InitConfig()

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	Execute()
}

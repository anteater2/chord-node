package main

import (
	"fmt"
	"log"

	"github.com/anteater2/chord-node/utils"

	"github.com/anteater2/chord-node/config"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := utils.LocalAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		fmt.Printf("%s\n", addr.String())
	}

	if config.Creator() {
		fmt.Printf("Creating network with %d addresses\n", config.MaxKey())
	} else {
		fmt.Printf("Connecting to network at %s\n", config.Introducer())
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

func main() {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		fmt.Printf("could not open configuration file %s: %s\n", *configFile, err)
	}

	_, err = config.New(f)
	if err != nil {
		fmt.Println(err)
	}

}

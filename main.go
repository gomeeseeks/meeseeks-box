package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
	"gitlab.com/mr-meeseeks/meeseeks-box/slack"
	"gitlab.com/mr-meeseeks/meeseeks-box/version"
)

func main() {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	showVersion := flag.Bool("version", false, "print the version and exit")

	flag.Parse()

	if *showVersion {
		log.Println(version.AppVersion)
		os.Exit(0)
	}
	if *debugMode {
		log.SetLevel(log.DebugLevel)
	}

	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatalf("could not open configuration file %s: %s\n", *configFile, err)
	}

	cnf, err := config.New(f)
	if err != nil {
		fmt.Println(err)
	}

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatalf("SLACK_TOKEN env var is empty")
	}

	client, err := slack.New(slack.ClientConfig{
		Token: token,
		Debug: *debugMode,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to slack")

	meeseek := meeseeks.New(client, cnf)
	ch := make(chan slack.Message)
	go client.ListenMessages(ch)
	for message := range ch {
		meeseek.Process(message)
	}
}

package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks"
	"gitlab.com/mr-meeseeks/meeseeks-box/slack"
)

func main() {
	// configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debug := flag.Bool("debug", false, "enabled debug mode")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// f, err := os.Open(*configFile)
	// if err != nil {
	// 	log.Fatalf("could not open configuration file %s: %s\n", *configFile, err)
	// }

	// _, err = config.New(f)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatalf("SLACK_TOKEN env var is empty")
	}

	client, err := slack.New(slack.ClientConfig{
		Token: token,
		Debug: *debug,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to slack")

	ch := make(chan slack.Message)
	go client.ListenMessages(ch)
	for m := range ch {
		meeseeks.ProcessMessage(m, client)
	}
}

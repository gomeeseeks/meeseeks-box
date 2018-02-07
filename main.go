package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomeeseeks/meeseeks-box/formatter"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/messenger"
	"github.com/gomeeseeks/meeseeks-box/slack"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	debugSlack := flag.Bool("debug-slack", false, "enabled debug mode for slack")
	showVersion := flag.Bool("version", false, "print the version and exit")
	apiAddress := flag.String("api-endpoint", ":9696", "api endpoint in which to listen for api calls")
	apiPath := flag.String("api-path", "/message", "api path in to listen for api calls")

	flag.Parse()

	if *showVersion {
		log.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	if *debugMode {
		log.SetLevel(log.DebugLevel)
	}

	cnf, err := config.LoadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := config.LoadConfig(cnf); err != nil {
		log.Fatalf("Could not load configuration: %s", err)
	}

	log.Info("Loaded configuration")

	slackClient, err := slack.Connect(*debugSlack, os.Getenv("SLACK_TOKEN"))
	if err != nil {
		log.Fatalf("Could not connect to slack: %s", err)
	}

	log.Info("Connected to slack")

	apiServer := api.NewServer(slackClient, *apiAddress)
	go func() {
		err = apiServer.ListenAndServe(*apiPath)
		if err != nil {
			log.Fatalf("Could not start API server: %s", err)
		}
	}()

	log.Infof("Started api server on %s%s", *apiAddress, *apiPath)

	msgs, err := messenger.Listen(slackClient, apiServer.GetListener())
	if err != nil {
		log.Fatalf("Could not initialize messenger subsystem: %s", err)
	}

	log.Info("Listening messages")

	meeseek := meeseeks.New(slackClient, msgs, formatter.New(cnf))
	go meeseek.Start()

	log.Info("Started commands pipeline")

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Listen for a signal forever
	sig := <-signalCh
	log.Infof("Got signal %s, trying to gracefully shutdown", sig)

	apiServer.Shutdown()
	msgs.Shutdown()
	meeseek.Shutdown()

	log.Infof("All done, quitting")
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/messenger"

	"github.com/pcarranza/meeseeks-box/meeseeks"
	"github.com/pcarranza/meeseeks-box/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	showVersion := flag.Bool("version", false, "print the version and exit")

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
	config.LoadConfig(cnf)

	messaging, err := messenger.Listen(messenger.MessengerOpts{
		Debug:      *debugMode,
		SlackToken: os.Getenv("SLACK_TOKEN"),
	})
	if err != nil {
		log.Fatalf("Could not initialize messenger subsystem: %s", err)
	}

	meeseek := meeseeks.New(messaging, cnf)
	go meeseek.Start(messaging.MessagesCh)

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Listen for a signal forever
	sig := <-signalCh
	log.Infof("Got signal %s, trying to gracefully shutdown", sig)

	messaging.Shutdown()
	meeseek.Shutdown()

	log.Infof("All done, quitting")
}

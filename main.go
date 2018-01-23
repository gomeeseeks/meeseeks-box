package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pcarranza/meeseeks-box/db"

	bolt "github.com/coreos/bbolt"
	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/config"
	"github.com/pcarranza/meeseeks-box/meeseeks"
	"github.com/pcarranza/meeseeks-box/slack"
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

	cnf, err := loadConfiguration(*configFile)
	must(err)

	auth.Configure(cnf)
	db.Configure(cnf)

	client, err := slack.Connect(*debugMode)
	must(err)

	log.Println("Connected to slack")

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	meeseek := meeseeks.New(client, cnf)
	go meeseek.Start()

	messages := make(chan slack.Message)
	go client.ListenMessages(messages)

processing:
	for {
		select {
		case sig := <-signalCh:
			log.Infof("Got signal %s, trying to gracefully shutdown", sig)
			close(messages)
			meeseek.Shutdown()
			break processing

		case message := <-messages:
			go func(message slack.Message) {
				meeseek.MessageCh <- message
			}(message)
		}
	}
	log.Infof("All done, quitting")
}

func loadConfiguration(configFile string) (config.Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return config.Config{}, fmt.Errorf("could not open configuration file %s: %s", configFile, err)
	}

	return config.New(f)
}

func must(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func openDB(cnf config.Config) (*bolt.DB, error) {
	db, err := bolt.Open(cnf.Database.Path, cnf.Database.Mode, &bolt.Options{
		Timeout: cnf.Database.Timeout,
	})
	db.Close()
	return db, err
}

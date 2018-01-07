package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
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

	cnf, err := loadConfiguration(*configFile)
	must(err)

	auth.Configure(cnf)

	client, err := slack.Connect(*debugMode)
	must(err)

	log.Println("Connected to slack")

	meeseek := meeseeks.New(client, cnf)

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	messages := make(chan slack.Message)

	go client.ListenMessages(messages)
	wg := sync.WaitGroup{}

processing:
	for {
		select {
		case sig := <-signalCh:
			log.Infof("Got signal %s, trying to gracefully shutdown", sig)
			close(messages)
			wg.Wait()
			break processing

		case message := <-messages:
			go func(message slack.Message) {
				wg.Add(1)
				defer wg.Done()

				meeseek.Process(message)
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

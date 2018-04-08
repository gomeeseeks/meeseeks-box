package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomeeseeks/meeseeks-box/formatter"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/jobs"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/messenger"
	"github.com/gomeeseeks/meeseeks-box/slack"
	"github.com/gomeeseeks/meeseeks-box/version"

	"github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	debugSlack := flag.Bool("debug-slack", false, "enabled debug mode for slack")
	showVersion := flag.Bool("version", false, "print the version and exit")
	apiAddress := flag.String("api-endpoint", ":9696", "api endpoint in which to listen for api calls")
	apiPath := flag.String("api-path", "/message", "api path in to listen for api calls")
	metricsPath := flag.String("metrics-path", "/metrics", "path to in which to expose prometheus metrics")
	slackStealth := flag.Bool("stealth", false, "Enable slack stealth mode")

	flag.Parse()

	if *showVersion {
		logrus.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	if *debugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}

	cnf, err := config.LoadFile(*configFile)
	if err != nil {
		logrus.Fatal(err)
	}
	if err := config.LoadConfig(cnf); err != nil {
		logrus.Fatalf("Could not load configuration: %s", err)
	}
	logrus.Info("Loaded configuration")

	if err := jobs.FailRunningJobs(); err != nil {
		logrus.Fatalf("Could not flush running jobs after: %s", err)
	}

	slackClient, err := slack.Connect(
		slack.ConnectionOpts{
			Debug:   *debugSlack,
			Token:   os.Getenv("SLACK_TOKEN"),
			Stealth: *slackStealth,
		})
	if err != nil {
		logrus.Fatalf("Could not connect to slack: %s", err)
	}
	logrus.Info("Connected to slack")

	httpServer := api.NewServer(slackClient, *metricsPath, *apiPath, *apiAddress)
	go func() {
		err = httpServer.ListenAndServe()
		if err != nil {
			logrus.Fatalf("Could not start API server: %s", err)
		}
	}()
	logrus.Infof("Started http server on %s%s", *apiAddress, *apiPath)

	msgs, err := messenger.Listen(slackClient, httpServer.GetListener())
	if err != nil {
		logrus.Fatalf("Could not initialize messenger subsystem: %s", err)
	}
	logrus.Info("Listening to slack messages")

	meeseek := executor.New(slackClient, msgs, formatter.New(cnf))
	go meeseek.Start()
	logrus.Info("Started commands pipeline")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Listen for a signal forever
	sig := <-signalCh
	logrus.Infof("Got signal %s, trying to gracefully shutdown", sig)

	httpServer.Shutdown()
	msgs.Shutdown()
	meeseek.Shutdown()

	logrus.Info("Everything has been shut down, bye bye!")
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence/jobs"
	"github.com/gomeeseeks/meeseeks-box/slack"
	"github.com/gomeeseeks/meeseeks-box/version"

	"github.com/sirupsen/logrus"
)

func main() {
	args := parseArgs()

	setLogLevel(args)
	loadConfiguration(args)

	cleanupPendingJobs()

	slackClient := connectToSlack(args)
	httpServer := listenHTTP(slackClient, args)

	logrus.Info("Listening to slack messages")

	exc := executor.New(slackClient)

	exc.ListenTo(slackClient)
	exc.ListenTo(httpServer)

	go exc.Run()
	logrus.Info("Started commands pipeline")

	waitForSignals(func() { // this locks for good, but receives a shutdown function
		httpServer.Shutdown()
		exc.Shutdown()
	})

	logrus.Info("Everything has been shut down, bye bye!")
}

type args struct {
	ConfigFile  string
	DebugMode   bool
	StealthMode bool
	DebugSlack  bool
	APIAddress  string
	APIPath     string
	MetricsPath string
	SlackToken  string
}

func parseArgs() args {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	debugSlack := flag.Bool("debug-slack", false, "enabled debug mode for slack")
	showVersion := flag.Bool("version", false, "print the version and exit")
	apiAddress := flag.String("api-endpoint", ":9696", "api endpoint in which to listen for api calls")
	apiPath := flag.String("api-path", "/message", "api path in to listen for api calls")
	metricsPath := flag.String("metrics-path", "/metrics", "path to in which to expose prometheus metrics")
	slackStealth := flag.Bool("stealth", false, "Enable slack stealth mode")
	slackToken := flag.String("slack-token", os.Getenv("SLACK_TOKEN"), "slack token, by default loaded from the SLACK_TOKEN environment variable")

	flag.Parse()

	if *showVersion {
		logrus.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	return args{
		ConfigFile:  *configFile,
		DebugMode:   *debugMode,
		StealthMode: *slackStealth,
		DebugSlack:  *debugSlack,
		SlackToken:  *slackToken,
		APIAddress:  *apiAddress,
		APIPath:     *apiPath,
		MetricsPath: *metricsPath,
	}
}

func setLogLevel(args args) {
	if args.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func loadConfiguration(args args) {
	cnf, err := config.LoadFile(args.ConfigFile)
	if err != nil {
		logrus.Fatal(err)
	}
	if err := config.LoadConfig(cnf); err != nil {
		logrus.Fatalf("Could not load configuration: %s", err)
	}
	logrus.Info("Configuration loaded")
}

func cleanupPendingJobs() {
	if err := jobs.FailRunningJobs(); err != nil {
		logrus.Fatalf("Could not flush running jobs after: %s", err)
	}
}

func connectToSlack(args args) *slack.Client {
	slackClient, err := slack.Connect(
		slack.ConnectionOpts{
			Debug:   args.DebugSlack,
			Token:   args.SlackToken,
			Stealth: args.StealthMode,
		})
	if err != nil {
		logrus.Fatalf("Could not connect to slack: %s", err)
	}
	logrus.Info("Connected to slack")

	return slackClient
}

func listenHTTP(client *slack.Client, args args) *api.Server {
	metrics.RegisterPath(args.MetricsPath)

	httpServer := api.NewServer(client, args.APIPath, args.APIAddress)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			logrus.Fatalf("Could not start HTTP server: %s", err)
		}
	}()
	logrus.Infof("Started HTTP server on %s", args.APIAddress)

	return httpServer
}

func waitForSignals(shutdownGracefully func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Listen for a signal forever
	sig := <-signalCh

	logrus.Infof("Got signal %s, shutting down gracefully", sig)

	shutdownGracefully()
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/http"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence"
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
	logrus.Info("Listening to slack messages")

	httpServer := listenHTTP(args)
	apiService := startAPI(slackClient, args)

	exc := executor.New(slackClient)

	exc.ListenTo(slackClient)
	exc.ListenTo(apiService)

	go exc.Run()
	logrus.Info("Started commands pipeline")

	waitForSignals(func() { // this locks for good, but receives a shutdown function
		httpServer.Shutdown()
		exc.Shutdown()
	})

	logrus.Info("Everything has been shut down, bye bye!")
}

type args struct {
	ConfigFile    string
	DebugMode     bool
	StealthMode   bool
	DebugSlack    bool
	Address       string
	APIPath       string
	MetricsPath   string
	SlackToken    string
	ExecutionMode string
	RemoteServer  string
}

func parseArgs() args {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	debugSlack := flag.Bool("debug-slack", false, "enabled debug mode for slack")
	showVersion := flag.Bool("version", false, "print the version and exit")
	address := flag.String("endpoint", ":9696", "http endpoint in which to listen")
	apiPath := flag.String("api-path", "/message", "api path in to listen for api calls")
	metricsPath := flag.String("metrics-path", "/metrics", "path to in which to expose prometheus metrics")
	slackStealth := flag.Bool("stealth", false, "Enable slack stealth mode")
	slackToken := flag.String("slack-token", os.Getenv("SLACK_TOKEN"), "slack token, by default loaded from the SLACK_TOKEN environment variable")
	executionMode := flag.String("mode", "standalone", "sets the execution mode, possible modes are standalone (default), server and agent")
	remoteServer := flag.String("server", "", "remote server to connect to, needed when executing in agent mode")

	flag.Parse()

	if *showVersion {
		logrus.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	return args{
		ConfigFile:    *configFile,
		DebugMode:     *debugMode,
		StealthMode:   *slackStealth,
		DebugSlack:    *debugSlack,
		SlackToken:    *slackToken,
		Address:       *address,
		APIPath:       *apiPath,
		MetricsPath:   *metricsPath,
		ExecutionMode: *executionMode,
		RemoteServer:  *remoteServer,
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
	if err := persistence.Jobs().FailRunningJobs(); err != nil {
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

func listenHTTP(args args) *http.Server {

	httpServer := http.New(args.Address)
	metrics.RegisterPath(args.MetricsPath)

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			logrus.Fatalf("Could not start HTTP server: %s", err)
		}
	}()
	logrus.Infof("Started HTTP server on %s", args.Address)

	return httpServer
}

func startAPI(client *slack.Client, args args) *api.Service {
	return api.New(client, args.APIPath)
}

func waitForSignals(shutdownGracefully func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Listen for a signal forever
	sig := <-signalCh

	logrus.Infof("Got signal %s, shutting down gracefully", sig)

	shutdownGracefully()
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/http"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/agent"
	"github.com/gomeeseeks/meeseeks-box/slack"
	"github.com/gomeeseeks/meeseeks-box/version"

	"github.com/sirupsen/logrus"
)

func main() {
	args := parseArgs()

	setLogLevel(args)
	httpServer := listenHTTP(args)

	shutdownFunc, err := launch(args)
	must("could not launch meeseeks-box: %s", err)

	waitForSignals(func() { // this locks for good, but receives a shutdown function
		httpServer.Shutdown()
		shutdownFunc()
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
	remoteServer := flag.String("server", "", "remote server to connect to, needed when executing in agent mode")

	flag.Parse()

	if *showVersion {
		logrus.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	var executionMode string
	if flag.NArg() == 0 {
		executionMode = "server"
	} else {
		executionMode = flag.Arg(1)
	}

	if !(executionMode == "server" || executionMode == "agent") {
		logrus.Println("Invalid execution mode. Valid modes are server (default), and agent")
		flag.Usage()
		os.Exit(1)
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
		ExecutionMode: executionMode,
		RemoteServer:  *remoteServer,
	}
}

func launch(args args) (func(), error) {
	switch args.ExecutionMode {
	case "server":
		loadConfiguration(args)
		cleanupPendingJobs()

		slackClient := connectToSlack(args)
		apiService := startAPI(slackClient, args)

		exc := executor.New(executor.Args{
			ConcurrentTaskCount: 20,
			WithBuiltinCommands: true,
			ChatClient:          slackClient,
		})

		exc.ListenTo(slackClient)
		exc.ListenTo(apiService)

		go exc.Run()

		return func() {
			exc.Shutdown()
		}, nil

	case "agent":
		remoteClient := agent.New(agent.Configuration{}) // This is deeply wrong and must be completed
		must("could not connect to remote server: %s", remoteClient.Connect())
		must("could not register and run this agent: %s", remoteClient.RegisterAndRun())

		return func() {
			remoteClient.Shutdown()
		}, nil

	default:
		return nil, fmt.Errorf("Invalid execution mode %s, Valid execution modes are server (default), and agent",
			args.ExecutionMode)
	}
}

func setLogLevel(args args) {
	if args.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func loadConfiguration(args args) {
	cnf, err := config.LoadFile(args.ConfigFile)
	must("could not load configuration file: %s", err)
	must("could not load configuration: %s", config.LoadConfig(cnf))

	logrus.Info("Configuration loaded")
}

func cleanupPendingJobs() {
	must("Could not flush running jobs after: %s", persistence.Jobs().FailRunningJobs())
}

func connectToSlack(args args) *slack.Client {
	slackClient, err := slack.Connect(
		slack.ConnectionOpts{
			Debug:   args.DebugSlack,
			Token:   args.SlackToken,
			Stealth: args.StealthMode,
		})

	must("Could not connect to slack: %s", err)
	logrus.Info("Connected to slack")

	return slackClient
}

func listenHTTP(args args) *http.Server {

	httpServer := http.New(args.Address)
	metrics.RegisterPath(args.MetricsPath)

	go func() {
		err := httpServer.ListenAndServe()
		must("Could not start HTTP server: %s", err)
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

func must(message string, err error) {
	if err != nil {
		logrus.Fatalf(message, err)
	}
}

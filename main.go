package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/config"
	"github.com/gomeeseeks/meeseeks-box/http"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/executor"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/agent"
	"github.com/gomeeseeks/meeseeks-box/remote/server"
	"github.com/gomeeseeks/meeseeks-box/slack"
	"github.com/gomeeseeks/meeseeks-box/version"

	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
)

func main() {
	args := parseArgs()

	setLogLevel(args)

	shutdownFunc, err := launch(args)
	must("could not launch meeseeks-box: %s", err)

	waitForSignals(func() { // this locks for good, but receives a shutdown function
		shutdownFunc()
	})

	logrus.Info("Everything has been shut down, bye bye!")
}

type args struct {
	ConfigFile        string
	DebugMode         bool
	StealthMode       bool
	DebugSlack        bool
	Address           string
	APIPath           string
	MetricsPath       string
	SlackToken        string
	ExecutionMode     string
	AgentOf           string
	GRPCServerAddress string
	GRPCServerEnabled bool
}

func parseArgs() args {
	configFile := flag.String("config", os.ExpandEnv("${HOME}/.meeseeks.yaml"), "meeseeks configuration file")
	debugMode := flag.Bool("debug", false, "enabled debug mode")
	debugSlack := flag.Bool("debug-slack", false, "enabled debug mode for slack")
	showVersion := flag.Bool("version", false, "print the version and exit")
	address := flag.String("http-address", ":9696", "http endpoint in which to listen")
	apiPath := flag.String("api-path", "/message", "api path in to listen for api calls")
	metricsPath := flag.String("metrics-path", "/metrics", "path to in which to expose prometheus metrics")
	slackStealth := flag.Bool("stealth", false, "Enable slack stealth mode")
	slackToken := flag.String("slack-token", os.Getenv("SLACK_TOKEN"), "slack token, by default loaded from the SLACK_TOKEN environment variable")
	agentOf := flag.String("agent-of", "", "remote server to connect to, enables agent mode")
	grpcServerAddress := flag.String("grpc-address", ":9697", "grpc server endpoint, used to connect remote agents")
	grpcServerEnabled := flag.Bool("with-grpc-server", false, "enable grpc remote server to connect to")

	flag.Parse()

	if *showVersion {
		logrus.Printf("Version: %s Commit: %s Date: %s", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	executionMode := "server"
	if *agentOf != "" {
		executionMode = "agent"
	}

	return args{
		ConfigFile:        *configFile,
		DebugMode:         *debugMode,
		StealthMode:       *slackStealth,
		DebugSlack:        *debugSlack,
		SlackToken:        *slackToken,
		Address:           *address,
		APIPath:           *apiPath,
		MetricsPath:       *metricsPath,
		AgentOf:           *agentOf,
		GRPCServerAddress: *grpcServerAddress,
		GRPCServerEnabled: *grpcServerEnabled,
		ExecutionMode:     executionMode,
	}
}

func launch(args args) (func(), error) {
	cnf, err := config.LoadFile(args.ConfigFile)
	must("failed to load configuration file: %s", err)

	switch args.ExecutionMode {
	case "server":
		must("could not load configuration: %s", config.LoadConfig(cnf))

		cleanupPendingJobs()

		httpServer := listenHTTP(args)
		remoteServer := startRemoteServer(args)

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
			httpServer.Shutdown()
			remoteServer.Shutdown()
		}, nil

	case "agent":
		remoteClient := agent.New(agent.Configuration{
			ServerURL:   args.AgentOf,
			Token:       "null-token",
			GRPCTimeout: 10 * time.Second,
			Commands:    cnf.Commands,
			Labels:      map[string]string{},
			// Options: add some options so we have at least some security, or at least make insecure optional
		})

		must("could not connect to remote server: %s", remoteClient.Connect())
		must("could not register and run this agent: %s", remoteClient.RegisterAndRun())

		logrus.Debugf("agent running connected to remote server: %s", args.AgentOf)

		return func() {
			remoteClient.Shutdown()
		}, nil

	default:
		return nil, fmt.Errorf("Invalid execution mode %s, Valid execution modes are server (default), and agent",
			args.ExecutionMode)
	}
}

func setLogLevel(args args) {
	logrus.AddHook(filename.NewHook())
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if args.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func cleanupPendingJobs() {
	must("Could not flush running jobs after: %s", persistence.Jobs().FailRunningJobs())
}

func connectToSlack(args args) *slack.Client {
	logrus.Debug("Connecting to slack")
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
	logrus.Debug("Creating a new http server")
	httpServer := http.New(args.Address)
	metrics.RegisterPath(args.MetricsPath)

	go func() {
		logrus.Debug("Listening on http")
		// err :=
		httpServer.ListenAndServe()
		// must("Could not start HTTP server: %s", err)
	}()
	logrus.Infof("Started HTTP server on %s", args.Address)

	return httpServer
}

func startAPI(client *slack.Client, args args) *api.Service {
	logrus.Debug("Starting api server")
	return api.New(client, args.APIPath)
}

func startRemoteServer(args args) *server.RemoteServer {
	s := server.New()
	if args.GRPCServerEnabled {
		logrus.Debugf("starting grpc remote server on %s", args.GRPCServerAddress)
		go func() {
			must("could not start grpc server", s.Listen(args.GRPCServerAddress))
		}()
	}

	return s
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
		os.Exit(1)
	}
}

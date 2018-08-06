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

	configureLogger(args)

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
	GRPCSecurityMode  string
	GRPCCertPath      string
	GRPCKeyPath       string
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

	grpcSecurityMode := flag.String("grpc-security-mode", "insecure", "grpc security mode, by default insecure, can be set to tls (for now)")
	grpcCertPath := flag.String("grpc-cert-path", "", "Cert to use with the GRPC server")
	grpcKeyPath := flag.String("grpc-key-path", "", "Key to use with the GRPC server")

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

		GRPCSecurityMode: *grpcSecurityMode,
		GRPCCertPath:     *grpcCertPath,
		GRPCKeyPath:      *grpcKeyPath,

		ExecutionMode: executionMode,
	}
}

func launch(args args) (func(), error) {
	cnf, err := config.ReadFile(args.ConfigFile)
	must("failed to load configuration file: %s", err)

	httpServer := listenHTTP(args)

	switch args.ExecutionMode {
	case "server":
		must("could not load configuration: %s", config.LoadConfiguration(cnf))
		must("Could not flush running jobs after: %s", persistence.Jobs().FailRunningJobs())

		metrics.RegisterServerMetrics()
		remoteServer, err := startRemoteServer(args)
		must("could not start GRPC server: %s", err)

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
		// metrics.RegisterAgentMetrics()

		remoteClient := agent.New(agent.Configuration{
			ServerURL:    args.AgentOf,
			Token:        "null-token",
			GRPCTimeout:  10 * time.Second,
			Commands:     cnf.Commands,
			Labels:       map[string]string{},
			SecurityMode: args.GRPCSecurityMode,
			CertPath:     args.GRPCCertPath,
		})

		must("could not connect to remote server: %s", remoteClient.Connect())

		go remoteClient.Run()

		logrus.Debugf("agent running connected to remote server: %s", args.AgentOf)

		return func() {
			remoteClient.Shutdown()
		}, nil

	default:
		return nil, fmt.Errorf("Invalid execution mode %s, Valid execution modes are server (default), and agent",
			args.ExecutionMode)
	}
}

func configureLogger(args args) {
	logrus.AddHook(filename.NewHook())
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if args.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
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
	httpServer := http.New(args.Address)
	metrics.RegisterPath(args.MetricsPath)

	go func() {
		logrus.Debug("Listening on http")
		httpServer.ListenAndServe()
	}()
	logrus.Infof("Started HTTP server on %s", args.Address)

	return httpServer
}

func startAPI(client *slack.Client, args args) *api.Service {
	logrus.Debug("Starting api server")
	return api.New(client, args.APIPath)
}

func startRemoteServer(args args) (*server.RemoteServer, error) {
	s, err := server.New(server.Config{
		CertPath:     args.GRPCCertPath,
		KeyPath:      args.GRPCKeyPath,
		SecurityMode: args.GRPCSecurityMode,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create GRPC Server: %s", err)
	}
	if args.GRPCServerEnabled {
		logrus.Debugf("starting grpc remote server on %s", args.GRPCServerAddress)
		go func() {
			must("could not start grpc server", s.Listen(args.GRPCServerAddress))
		}()
	}

	return s, nil
}

func waitForSignals(shutdownGracefully func()) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

Loop:
	for sig := range signalCh { // Listen for a signal forever
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			logrus.Infof("Got signal %s, shutting down gracefully", sig)
			break Loop

		case syscall.SIGHUP:
			// reload configuration

		case syscall.SIGUSR1:
			toggleDebugLogging()
		}
	}

	shutdownGracefully()
}

func toggleDebugLogging() {
	switch logrus.GetLevel() {
	case logrus.DebugLevel:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func must(message string, err error) {
	if err != nil {
		logrus.Fatalf(message, err)
		os.Exit(1)
	}
}

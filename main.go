package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"

	"github.com/magdyamr542/service-monitorer/config"
	"github.com/magdyamr542/service-monitorer/http"
	"github.com/magdyamr542/service-monitorer/informer"
	"github.com/magdyamr542/service-monitorer/monitorer"
	"gopkg.in/yaml.v2"
)

// Vars injected by the build command
var (
	GitCommit = ""
)

var (
	logLevel = map[string]log.Level{
		"debug": log.DebugLevel,
		"info":  log.InfoLevel,
		"error": log.ErrorLevel,
	}
)

func main() {

	configFile := flag.String("config", "config.yaml", "path to config file")
	loglevel := flag.String("loglevel", "debug", "log level. one of (debug,info,error)")
	flag.Parse()

	logger := log.New(os.Stdout)
	logger.SetPrefix("[service-monitorer]")

	if _, ok := logLevel[*loglevel]; !ok {
		logger.Fatalf("Bad log level")
	}
	logger.SetLevel(logLevel[*loglevel])

	logger.Infof("Version: %s", GitCommit)

	yamlFile, err := os.ReadFile(*configFile)
	if err != nil {
		logger.Fatalf("Readfile: %v", err)
	}

	var c config.Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logger.Fatalf("Unmarshal: %v", err)
	}

	if err := c.Validate(); err != nil {
		logger.Fatalf("config invalid: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		logger.Warnf("Context cancelled. Exiting.")
		cancel()
	}()

	httpClient := http.NewClient()
	informers := map[informer.SupportedInformer]informer.Informer{
		informer.Slack: informer.NewSlack(logger, httpClient),
	}
	for _, key := range informer.SupportedInformers {
		if _, ok := informers[key]; !ok {
			logger.Fatalf("Informer %s not registered. Can't start the service", key)
		}
	}

	// Start monitoring...
	monitorer := monitorer.NewMonitorer(c, httpClient, informers, logger)
	if err := monitorer.Monitor(ctx); err != nil {
		log.Fatalf("Monitorer: %v", err)
	}

}

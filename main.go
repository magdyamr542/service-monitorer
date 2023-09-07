package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/magdyamr542/service-monitorer/config"
	"gopkg.in/yaml.v2"
)

// Vars injected by the build command
var (
	GitCommit = ""
)

func main() {
	logger := log.New(os.Stdout, "[service-monitorer] ", log.Ldate|log.LUTC|log.Ltime)
	logger.Printf("Version: %s\n", GitCommit)

	configFile := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	yamlFile, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Readfile: %v", err)
	}

	var c config.Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	if err := c.Validate(); err != nil {
		fmt.Printf("config invalid: %v\n", err)
	}
}

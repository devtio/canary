// test.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/devtio/canary/config"
	"github.com/devtio/canary/log"
	server "github.com/devtio/canary/server"
	"github.com/devtio/canary/status"

	"github.com/golang/glog"
)

var (
	version    = "unknown"
	commitHash = "unknown"
)

// Command line arguments
var (
	argConfigFile = flag.String("config", "", "Path to the YAML configuration file. If not specified, environment variables will be used for configuration.")
)

func main() {
	defer glog.Flush()

	// process command line
	flag.Parse()
	validateFlags()

	// log startup information
	log.Infof("Devtio: Version: %v, Commit: %v\n", version, commitHash)
	log.Debugf("Devtio: Command line: [%v]", strings.Join(os.Args, " "))

	// load config file if specified, otherwise, rely on environment variables to configure us
	if *argConfigFile != "" {
		c, err := config.LoadFromFile(*argConfigFile)
		if err != nil {
			glog.Fatal(err)
		}
		config.Set(c)
	} else {
		log.Infof("No configuration file specified. Will rely on environment for configuration.")
		config.Set(config.NewConfig())
	}
	log.Tracef("Devtio Configuration:\n%s", config.Get())

	if err := validateConfig(); err != nil {
		glog.Fatal(err)
	}

	status.Put(status.CoreVersion, version)
	status.Put(status.CoreCommitHash, commitHash)

	// Start listening to requests
	server := server.NewServer()
	server.Start()

	// wait forever, or at least until we are told to exit
	waitForTermination()

	// Shutdown internal components
	log.Info("Shutting down internal components")
	server.Stop()
}

func waitForTermination() {
	// Channel that is notified when we are done and should exit
	// TODO: may want to make this a package variable - other things might want to tell us to exit
	var doneChan = make(chan bool)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			log.Info("Termination Signal Received")
			doneChan <- true
		}
	}()

	<-doneChan
}

func validateConfig() error {
	if config.Get().Server.Port < 0 {
		return fmt.Errorf("server port is negative: %v", config.Get().Server.Port)
	}

	if err := config.Get().Server.Credentials.ValidateCredentials(); err != nil {
		return fmt.Errorf("server credentials are invalid: %v", err)
	}
	return nil
}

func validateFlags() {
	if *argConfigFile != "" {
		if _, err := os.Stat(*argConfigFile); err != nil {
			if os.IsNotExist(err) {
				log.Debugf("Configuration file [%v] does not exist.", *argConfigFile)
			}
		}
	}
}

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/natemarks/preflight/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	version string = "v0.0.9"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)

	// I tried to move this to init() but it doesn't work there
	log.SetOutput(os.Stdout)
	config.GetSettings()
	RealMain()
}

func RealMain() {
	// init success to true.  any failing check with set it to false
	var success bool = true
	verbose, err := strconv.ParseBool(viper.GetString("verbose"))
	if err != nil {
		panic("Unable to get config key: verbose")
	}
	if verbose {
		log.SetLevel(log.TraceLevel)
		log.Debug("Verbose logging is enabled")
	}
	log.Info(fmt.Sprintf("preflight version: %s", version))

	config.LogContainerMetadata()

	// get the list of environment variables the service nees so we can check them
	EnvVarsToCheck := viper.GetStringSlice("checked_environment_variables")
	if len(EnvVarsToCheck) == 0 {
		success = false
		msg := "Unable to get a list of environment variables to check. set 'checked_environment_variables' in the config"
		log.Error(msg)
	}

	// make sure each of the required env vars has some set value
	varMap, ok := config.CheckVars(EnvVarsToCheck)
	if !ok {
		success = false
		log.Error("Some required environment variables were not set")
	}

	// some  env vars might have data relevant to host checks.  capture that data into a map of host maps by ID
	hostMap := config.GetHosts(varMap)

	// temporarily drop the reachableHosts variable to run tests
	// reachableHosts, err := config.GetReachableHosts(hostMap)
	_, ok = config.GetReachableHosts(hostMap)
	if !ok {
		success = false
		log.Error("Some hosts are not reachable")
	}

	// success was initialized to true. Ay failing test would have set it to false
	if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}

}

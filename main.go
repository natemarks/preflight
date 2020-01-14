package main

import (
	"os"
	"strconv"

	"github.com/natemarks/preflight/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	verbose, err := strconv.ParseBool(viper.GetString("verbose"))
	if err != nil {
		panic("Unable to get config key: verbose")
	}
	if verbose {
		log.SetLevel(log.TraceLevel)
		log.Debug("Verbose logging is enabled")
	}
	err = config.CheckVars(viper.GetStringSlice("checked_environment_variables"))
	{
		if err != nil {
			log.Error("Unable to check environment variables")
			os.Exit(2)
		}
	}

}

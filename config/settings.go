package config

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	DefaultVerbose bool = false
)

// Get all of the config settings from file, environment, flag, etc and return a config object
func GetSettings() {
	DefineViperDefaults()
	DefineViperConfigFile()
	// read in the config file because it contains the env vars we need to scan
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Warn("No config file found")
	}

	DefineViperEnvVars()

}

func DefineViperDefaults() {
	viper.SetDefault("verbose", DefaultVerbose)
}

func DefineViperConfigFile() {
	viper.SetConfigName("preflight") // name of config file (without extension)
	viper.AddConfigPath("testdata/") // check testdata first
	viper.AddConfigPath(".")         // then look in the same dir
}

func DefineViperEnvVars() {
	viper.SetEnvPrefix("pf") // will be uppercased automatically
	ll := viper.GetStringSlice("environment_variables")

	for _, ev := range ll {
		err := viper.BindEnv(ev)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to bind environment variable: %s", ev))
		}
	}
}

package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

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
	DefineViperEnvVars()

	// read in the config file because it contains the env vars we need to scan
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Warn("No config file found")
	}

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
	viper.SetEnvPrefix("pf") // will be uppercase automatically
	viper.AutomaticEnv()
}

func GetHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

type Connection struct {
	id, address, port, username, password, token, description, version string
}

// Return true if the environment variable is set to a non-empty value
func IsSet(ev string) error {
	name, ok := os.LookupEnv(ev)
	if ok {
		if os.Getenv(ev) == "" {
			errorMsg := fmt.Sprintf("environment variable set, but empty: %s", name)
			log.Error(errorMsg)
			return errors.New(errorMsg)
		} else {
			values := []interface{}{ev, GetHash(os.Getenv(name))}
			log.Info(fmt.Sprintf("environment variable found: %s = %s (sha256)", values...))
			return nil
		}
	} else {
		errorMsg := fmt.Sprintf("environment variable key does not exist: %s", name)
		log.Error(errorMsg)
		return errors.New(errorMsg)
	}
}

func CheckVars() {
	ll := viper.GetStringSlice("checked_environment_variables")
	for _, vv := range ll {
		err := IsSet(vv)
		if err != nil {
			break
		}
	}
	log.Info(fmt.Sprintf("Checked %d environment variables.  Finished", len(ll)))
}

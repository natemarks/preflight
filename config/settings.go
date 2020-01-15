package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/natemarks/preflight/utility"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	DefaultVerbose bool = false
)

var ReservedSuffixes = []string{
	"USERNAME",
	"PASSWORD",
	"TOKEN",
	"DESCRIPTION",
	"VERSION",
	"TIMEOUT",
}

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

// Check a list of environment variables, stopping onf the first failure
func CheckVars(ll []string) error {
	if len(ll) == 0 {
		return errors.New("no environment variables to check")
	}
	for _, vv := range ll {
		err := IsSet(vv)
		if err != nil {
			break
		}
	}
	log.Info(fmt.Sprintf("Checked %d environment variables.  Finished", len(ll)))
	return nil
}

// Return a list of Connections from the list of environment variables
// Based on reserved suffixes (ex. username, password, token, etc)
// APPLES_DATABASE_USERNAME='apple_user'
// the connection id would be "APPLES_DATABASE" and the username would be 'apple_user'
// match without case, but the environment variables should be caps
func Connections(ll []string) (map[string]map[string]string, error) {
	var ccs = map[string]map[string]string{}
	for _, vv := range ll {
		words := strings.Split(vv, "_")
		suffix := words[len(words)-1]
		if !utility.Contains(ReservedSuffixes, strings.ToUpper(suffix)) {
			continue
		} else {
			id := strings.TrimSuffix(vv, "_"+suffix)
			ccs[id] = make(map[string]string)
			ccs[id][suffix] = os.Getenv(vv)
		}
	}
	return ccs, nil
}

// given either a cidr or a host name, return the IP  address or error out
// www.google.com -> 1.2.3.4
// 1.2.3.4 -> 1.2.3.4
func ResolveHostName(hn string) (string, error) {
	//try to parse the string as an IP address
	i, _, err := net.ParseCIDR(hn)
	// if it can't be parsed as a cidr, try to  resolve it as a host name
	if err != nil {
		log.Debug(fmt.Sprintf("%s is not an IP address. Resolving hostname", hn))
		lh, err := net.LookupHost(hn)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to resolve host: %s", hn))
		} else {
			log.Debug(fmt.Sprintf("Resolved %s to %s", hn, lh[0]))
		}
		return lh[0], err

	} else {
		return string(i), nil
	}
}

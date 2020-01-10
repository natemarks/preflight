package config

import (
	"crypto/sha256"
	"fmt"

	"github.com/natemarks/preflight/utility"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	DefaultVerbose bool = false
)

var ValidCredentialFields = []string{"username", "password", "token", "description", "version"}

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
	viper.SetEnvPrefix("pf") // will be uppercased automatically
	viper.AutomaticEnv()
}

func GetHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

type Host struct {
	address, port string
	credentials   []map[string]string
}

func NewHost(address, port string) *Host {
	h := new(Host)
	h.address = address
	h.port = port
	return h
}

// Filter and validate the credential map
func (h *Host) AddCredential(cred map[string]string) error {
	//create new keys for the hashed values of these fields
	hashedFields := []string{"username", "password", "token"}
	newCredentialMap := make(map[string]string)
	for k, v := range cred {
		if !utility.Contains(ValidCredentialFields, k) {
			log.Warn(fmt.Sprintf("Ignoring invalid credential field: %s", k))
			continue
		}
		newCredentialMap[k] = v
		if utility.Contains(hashedFields, k) {
			newCredentialMap[fmt.Sprintf("%sHashed", k)] = GetHash(v)
		}
	}
	h.credentials = append(h.credentials, newCredentialMap)

	return nil
}

package config

import (
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func TestORFromFile(t *testing.T) {
	GetSettings()

	// the default value is false
	// This checks for the value set in config/testdata/preflight.yaml
	if viper.GetString("Verbose") != "true" {
		t.Fail()
	}
	//This is has no default and is only set in the config file
	if viper.GetString("fileonly") != "unregistered value" {
		t.Fail()
	}
}

func TestORFromEV(t *testing.T) {
	var err error
	// Set some environment variables
	m := map[string]string{
		"PF_DB_PORT":     "2345",
		"PF_DB_HOSTNAME": "pg.domain.com",
	}
	for k, v := range m {
		err = os.Setenv(k, v)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to get environment variable: %s", "PF_DB_PORT"))
		}
	}
	// This get forces the lookup of envirnmnet variables
	GetSettings()
	var ll string
	ll = viper.GetString("db_port")
	if ll != "2345" {
		t.Fail()
	}

}

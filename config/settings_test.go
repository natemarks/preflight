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
	err := os.Setenv("PF_DB_PORT", "2345")
	if err != nil {
		log.Error(fmt.Sprintf("Unable to get environment variable: PF_DB_PORT"))
	}

	// This get forces the lookup of environment variables
	GetSettings()
	var ll string
	ll = viper.GetString("db_port")
	if ll != "2345" {
		t.Fail()
	}
}

func TestHost_AddCredential(t *testing.T) {
	h := NewHost("pg.domain.com", "2345")
	cm := make(map[string]string)
	for _, f := range ValidCredentialFields {
		cm[f] = fmt.Sprintf("worst_%s_ever", f)
	}
	h.AddCredential(cm)
	if len(h.credentials) != 1 {
		t.Fail()
	}

}

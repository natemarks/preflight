package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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

func TestGetHash(t *testing.T) {
	got := GetHash("the quick brown...")
	if got != "d0d51358d9a060c093d77a9ab57469fc272a100a341dacf2332d798522ccf32e" {
		t.Fail()
	}
}

// Verify returned error and log message
func TestIsSetUnset(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Unsetenv("NONEXISTENT_VAR")
	err := IsSet("NONEXISTENT_VAR")

	// should log the non-existent variable message
	if !strings.Contains(hook.LastEntry().Message, "environment variable key does not exist:") {
		t.Fail()
	}
	// should return an error
	if err == nil {
		t.Fail()
	}
}

// Verify returned error and log message
func TestIsSetEmpty(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Setenv("EMPTY_VAR", "")
	err := IsSet("EMPTY_VAR")

	// should log the empty variable message
	if !strings.Contains(hook.LastEntry().Message, "environment variable set, but empty:") {
		t.Fail()
	}
	// should return an error
	if err == nil {
		t.Fail()
	}

}

// Verify returned error and log message
func TestIsSetSuccess(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Setenv("VALID_VAR", "VALID_VALUE")
	err := IsSet("VALID_VAR")

	// should log the success message
	if !strings.Contains(hook.LastEntry().Message, "environment variable found:") {
		t.Fail()
	}
	// should return nil error
	if err != nil {
		t.Fail()
	}
}

func TestConnections(t *testing.T) {
	_ = os.Setenv("MY_COOL_DB_USERNAME", "VALID_VALUE")
	ll := []string{"MY_COOL_DB_USERNAME"}
	got, _ := Connections(ll)
	if got["MY_COOL_DB"]["USERNAME"] != "VALID_VALUE" {
		t.Fail()
	}

}

func TestResolveHostNameIP(t *testing.T) {
	i, err := ResolveHostName("8.8.8.8")
	if err != nil {
		t.Fail()
	}
	if i != "8.8.8.8" {
		t.Fail()
	}
}

func TestResolveHostNameHost(t *testing.T) {
	i, err := ResolveHostName("www.google.com")
	if err != nil {
		t.Fail()
	}
	if len(i) == 0 {
		t.Fail()
	}
}

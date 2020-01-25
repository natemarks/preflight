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
	_, ok := IsSet("NONEXISTENT_VAR")

	// should log the non-existent variable message
	if !strings.Contains(hook.LastEntry().Message, "environment variable key does not exist:") {
		t.Fail()
	}
	// should return an error
	if ok {
		t.Fail()
	}
}

// Verify returned error and log message
func TestIsSetEmpty(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Setenv("EMPTY_VAR", "")
	_, ok := IsSet("EMPTY_VAR")

	// should log the empty variable message
	if !strings.Contains(hook.LastEntry().Message, "environment variable set, but empty:") {
		t.Fail()
	}
	// should return an error
	if ok {
		t.Fail()
	}

}

// Verify returned error and log message
func TestIsSetSuccess(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Setenv("VALID_VAR", "VALID_VALUE")
	_, ok := IsSet("VALID_VAR")

	// should log the success message
	if !strings.Contains(hook.LastEntry().Message, "environment variable found:") {
		t.Fail()
	}
	// should return nil error
	if !ok {
		t.Fail()
	}
}

// Given an IP address, return the IP address without errors
func TestResolveHostNameIP(t *testing.T) {
	i, ok := ResolveHostName("8.8.8.8")
	if !ok {
		t.Fail()
	}
	if i != "8.8.8.8" {
		t.Fail()
	}
}

// Given a valid host name, resolve and return the IP
func TestResolveHostNameHost(t *testing.T) {
	i, ok := ResolveHostName("www.google.com")
	if !ok {
		t.Fail()
	}
	if len(i) == 0 {
		t.Fail()
	}
}

func TestCanConnectSucceed(t *testing.T) {
	if !CanConnect("www.google.com", "80", 3000) {
		t.Fail()
	}
}

// Throw a mix of good and bad host maps into GetReachableHosts
// Only one host is good so the length of res should be 1
// because some entries failed, ok should be false
func TestGetReachableHosts(t *testing.T) {
	var testMap = map[string]map[string]string{}

	testMap["google"] = map[string]string{}
	testMap["unresolveable"] = map[string]string{}
	testMap["badport"] = map[string]string{}

	testMap["google"]["id"] = "google"
	testMap["google"]["address"] = "www.google.com"
	testMap["google"]["port"] = "80"

	testMap["unresolveable"]["id"] = "unresolveable"
	testMap["unresolveable"]["address"] = "unreasolvable.name.garbagetld"
	testMap["unresolveable"]["port"] = "80"

	testMap["badport"]["id"] = "badport"
	testMap["badport"]["address"] = "www.google.com"
	testMap["badport"]["port"] = "40444"

	res, ok := GetReachableHosts(testMap)
	if len(res) != 1 {
		t.Fail()
	}
	if ok {
		t.Fail()
	}

}

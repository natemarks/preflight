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

	got = GetHash("garbage")
	if got != "795b6904e54f82411df4b0e27a373a55eea3f9d66dac5a9bce1dd92f7b401da5" {
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

func TestCheckVarsEmpty(t *testing.T) {
	hook := test.NewGlobal()
	v := []string{}
	vMap, ok := CheckVars(v)
	if ok {
		t.Fail()
	}
	if len(vMap) != 0 {
		t.Fail()
	}
	if !strings.Contains(hook.Entries[0].Message, "no environment variables to check") {
		t.Fail()
	}

}

func TestCheckVarsValid(t *testing.T) {
	hook := test.NewGlobal()

	_ = os.Setenv("DEPLOYMENT_COLOR", "RED")
	// confusing entry that might kinda look like ust because of the USERNAME suffix
	// but doesn't begin with a client type
	_ = os.Setenv("SOME_NON_HOST_BUNCH_OF_STUFF_USERNAME", "garbage")
	// set some host stuff for the MY_EXPIRED_IDENTITES db
	_ = os.Setenv("POSTGRES10_MY_EXPIRED_IDENTITIES_USERNAME", "jdoe")
	_ = os.Setenv("POSTGRES10_MY_EXPIRED_IDENTITIES_PASSWORD", "bad_password")
	_ = os.Setenv("POSTGRES10_MY_EXPIRED_IDENTITIES_ADDRESS", "db.domain.invalid_tld")
	_ = os.Setenv("POSTGRES10_MY_EXPIRED_IDENTITIES_PORT", "5432")

	v := []string{
		"DEPLOYMENT_COLOR",
		"SOME_NON_HOST_BUNCH_OF_STUFF_USERNAME",
		"POSTGRES10_MY_EXPIRED_IDENTITIES_USERNAME",
		"POSTGRES10_MY_EXPIRED_IDENTITIES_PASSWORD",
		"POSTGRES10_MY_EXPIRED_IDENTITIES_ADDRESS",
		"POSTGRES10_MY_EXPIRED_IDENTITIES_PORT",
	}

	vMap, ok := CheckVars(v)
	if !ok {
		t.Fail()
	}
	if len(vMap) != 6 {
		t.Fail()
	}

	if hook.Entries[0].Message !=
		"environment variable found: DEPLOYMENT_COLOR = "+
			"65cbe1e19791d49b023800b3e22714e2589b6339bef5e660b628dd1023506ebd (sha256)" {
		t.Fail()
	}

	if hook.Entries[1].Message !=
		"environment variable found: SOME_NON_HOST_BUNCH_OF_STUFF_USERNAME = "+
			"795b6904e54f82411df4b0e27a373a55eea3f9d66dac5a9bce1dd92f7b401da5 (sha256)" {
		t.Fail()
	}

	if hook.Entries[2].Message !=
		"environment variable found: POSTGRES10_MY_EXPIRED_IDENTITIES_USERNAME = "+
			"d30a5f57532a603697ccbb51558fa02ccadd74a0c499fcf9d45b33863ee1582f (sha256)" {
		t.Fail()
	}

	if hook.Entries[3].Message !=
		"environment variable found: POSTGRES10_MY_EXPIRED_IDENTITIES_PASSWORD = "+
			"4323203b4ca7f7d3a7fc29d6f36a5b225e1293b718425fae911e7f078acbbf41 (sha256)" {
		t.Fail()
	}

	if hook.Entries[4].Message !=
		"environment variable found: POSTGRES10_MY_EXPIRED_IDENTITIES_ADDRESS = "+
			"3805b0bdbb9a90df54b68ca4897791462a2a1f069f78a3d297ce7092fcf8b49b (sha256)" {
		t.Fail()
	}

	if hook.Entries[5].Message !=
		"environment variable found: POSTGRES10_MY_EXPIRED_IDENTITIES_PORT = "+
			"4aeb7ad6d5d37a041c4c5ce6562bf9e3caf05a42d931cef4d9e2a60ca623194d (sha256)" {
		t.Fail()
	}

	if hook.Entries[6].Message != "Checked 6 environment variables.  Finished" {
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

// need something for this. maybe reuse the data from the CheckVars test
// write ths test to make suse I can remove the limitation from the readme
func TestGetHosts(t *testing.T) {

}

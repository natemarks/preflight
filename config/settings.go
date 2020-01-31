package config

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/natemarks/preflight/utility"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//  Host Data Filter Pipeline
// Many of the functions that check hosts take a nested map of host data organized by host id . in yaml it would look
// like:

//"IDENTITIES":
//	"ADDRESS": "iddy.domain.com"
//	"PORT": "5432"
//	"ID": "IDENTITIES"
//	"USERNAME": "jdoe"
//"HISTORY":
//	"ADDRESS": "hizzy.domain.com"
//	"PORT": "5432"
//	"ID": "HISTORY"
//	"USERNAME": "jdoe"

// Such functions iterate through the hosts, and return the map with the failing host checks filtered out AND a
// boolean 'ok' value that's only true if all the host checks passed.  In cases where subsequent tests are cheap, we
// use the unfiltered list to expose as many problems as possible as early as possible.  In cases where the there's no
// hope of the check passing and/or the test that's expected to fail will be resource expensive, we'd us the filtered
// list.

const (
	DefaultVerbose      bool   = false
	DefaultOrganization string = "MyCompanyName"
	DefaultTeam         string = "DevOps"
	EVWordSeparator     string = "_"
	ConnTimeoutMS       int64  = 3000 // default connection timeout in milliseconds
)

//  Supported clients is used to determine environment variables that contain host attributes AND
// to match a map host attributes to a client access testing function that can handle them
var SupportedClients = []string{
	"POSTGRES10",
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
	viper.SetDefault("organization", DefaultOrganization)
	viper.SetDefault("team", DefaultTeam)
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

//  BEGIN HERE

// Verify Environment Variables are set
// using the smaller set of verified EVs, find the EVs that contain connection data (by matching prefix to reserved clients)
// Find EVs that contain connection data, group them bu host ID, and merge all the maps for each host id together
//iterate through the host maps
// for each host map:
//  if IsRerachable:
//     VerifyClientACCESS

// MAIN:  Calls CheckVars first and saves the output to varMap

// Return a verified map of environment variables and values
func CheckVars(ll []string) (map[string]string, bool) {
	success := true
	res := make(map[string]string)
	if len(ll) == 0 {
		log.Error("no environment variables to check")
		success = false
	}
	for _, key := range ll {
		val, ok := IsSet(key)
		if ok {
			res[key] = val

		} else {
			success = false
			continue
		}
	}
	log.Info(fmt.Sprintf("Checked %d environment variables.  Finished", len(ll)))
	return res, success
}

// Return true if the environment variable is set to a non-empty value
func IsSet(key string) (string, bool) {
	success := true
	val, ok := os.LookupEnv(key)
	if ok {
		if val == "" {
			errorMsg := fmt.Sprintf("environment variable set, but empty: %s", val)
			log.Error(errorMsg)
			success = false
		} else {
			hash := GetHash(val)
			log.Info(fmt.Sprintf("environment variable found: %s = %s (sha256)", key, hash))
		}
	} else {
		errorMsg := fmt.Sprintf("environment variable key does not exist: %s", key)
		log.Error(errorMsg)
		success = false
	}
	return val, success
}

// MAIN calls GetHosts(varMap) and saves the returned map to hostMap

// Use preflight naming rules to generate a map of hosts by id
// Given a valid map of environment variable keys and values (presumably validated by CheckVars), return a map of hosts,
// by host id.
// not all environment variables contain host attributes. host data EVs have 3 or more parts and begin with a string
// that matches a supported client type
// host attributes are gathered by grouing and perging fdata from a number of environment varaiable example:

// to get this
// {
//	"ID": "HOT_PICKLES",
//	"CLIENT": "POSTGRES10",
//	"ADDRESS": "8.8.8.8",
//	"PORT": "5432",
// }

//  I would use
// POSTGRES10_HOT_PICKLES_ADDRESS=8.8.8.8
// POSTGRES10_HOT_PICKLES_PORT=5432

// These start as two maps that are merged into one:
// {
//	"ID": "HOT_PICKLES",
//	"CLIENT": "POSTGRES10",
//	"ADDRESS": "8.8.8.8",
// }
// {
//	"ID": "HOT_PICKLES",
//	"CLIENT": "POSTGRES10",
//	"PORT": "5432",
// }

func GetHosts(envVars map[string]string) map[string]map[string]string {
	res := make(map[string]map[string]string)

	for key, val := range envVars {
		thisEVMap, ok := GetHostFromEV(key, val)
		if !ok {
			continue
		} else {
			if val, ok := res[thisEVMap["ID"]]; ok {
				// the id exists. merge the map for this ev into the existing map
				UpdateMap(val, thisEVMap)
			} else {
				// if not create a new key and
				res[thisEVMap["ID"]] = thisEVMap
			}
		}

	}
	return res

}

// given a properly formatted environment variable key and it's value return a pointer to a host
// give a key and value: POSTGRES10_HOT_PICKLES_USERNAME=jdoe
// return a map like:
// {
//	"ID": "HOT_PICKLES",
//	"CLIENT": "POSTGRES10",
//	"username": "jdoe",
// }
// The first part is always the client. The last part is the field. All the midde parts are the identity
// If the first part matches a reserved string that represents a client type we can test, the environment variable is assumed to hold
// host connection information
func GetHostFromEV(key string, value string) (map[string]string, bool) {
	success := true
	words := strings.Split(key, EVWordSeparator)
	res := make(map[string]string)
	if len(words) < 3 {
		errMsg := fmt.Sprintf("too few fields in key to be a host setting: %s", key)
		success = false
		log.Debug(errMsg)
		return res, success
	}

	// strip the first slice entry out for the client and keep the remaining list in theRest
	client, theRest := words[0], words[1:]

	if !utility.Contains(SupportedClients, client) {
		errMsg := fmt.Sprintf("prefix doesn't match a supported client: %s", client)
		success = false
		log.Debug(errMsg)
		return res, success
	}

	//strip the last entry as the field name and keep the middle entries together
	fieldName, theMiddle := theRest[len(theRest)-1], theRest[:len(theRest)-1]

	//Join the middle values together into an id that can contain separators
	id := strings.Join(theMiddle, EVWordSeparator)
	res["ID"] = id
	res["CLIENT"] = client
	res[fieldName] = value
	return res, success
}

// write newMap keys and values into myMap
// Note: a map is always a reference so I don't have to return
func UpdateMap(myMap map[string]string, addMap map[string]string) {
	for k, v := range addMap {
		myMap[k] = v
	}
}

// Return map with failed checked filtered out and a boolean that's only true if everything succeeded
// See Host Data Filter Pipeline at the top for more information
func GetReachableHosts(hosts map[string]map[string]string) (map[string]map[string]string, bool) {
	success := true

	res := make(map[string]map[string]string)

	for _, hMap := range hosts {
		var ok bool
		hMap["ADDRESS"], ok = ResolveHostName(hMap["ADDRESS"])
		if !ok {
			success = false
			continue
		} else {
			ok := CanConnect(hMap["ADDRESS"], hMap["PORT"], ConnTimeoutMS)
			if ok {
				res[hMap["ID"]] = hMap
			} else {
				success = false
			}
		}

	}
	return res, success
}

// given either a cidr or a host name, return the IP  address or error out
// www.google.com -> 1.2.3.4
// 1.2.3.4 -> 1.2.3.4
func ResolveHostName(hn string) (string, bool) {
	success := true
	//try to parse the string as an IP address
	i, _, err := net.ParseCIDR(hn)
	// if it can't be parsed as a cidr, try to  resolve it as a host name
	if err != nil {
		log.Debug(fmt.Sprintf("%s is not an IP address. Resolving hostname", hn))
		lh, err := net.LookupHost(hn)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to resolve host: %s", hn))
			success = false
			return "", success
		} else {
			log.Debug(fmt.Sprintf("Resolved %s to %s", hn, lh[0]))
			return lh[0], success
		}
	} else {
		return string(i), success
	}
}

func CanConnect(address, port string, timeout int64) bool {
	target := address + ":" + port
	var success = true
	conn, err := net.DialTimeout("tcp", target, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to connect to  %s", target))
		success = false
	} else {
		log.Debug(fmt.Sprintf("Successfully connected to   %s", target))
		defer func() {
			err := conn.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

	}
	return success
}

func LogContainerMetadata() {
	///if it's AWS log the image name/version/whatever else we want from the localhost metadata json curl
}

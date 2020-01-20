# Preflight

preflight is run inside a docker container.  The container starts by running an entrypoint.sh script that looks like the one below. The preflight utility will run through every check and log the results to stdout which should get be readily available to any docker logger.  If any of the checks failed, it exits with a non-zero exit code.  The entrypoint.sh script must run with 'set -e' in order to be sure it exits if preflight throws an error *BEFORE* starting the service. This should achieve the following goals:
 
 - Log a thorough list of all the the missing/failing configuration settings adn tag the logs for the audience that needs to resolve them ex: \[MyOrgName:DevOps\] 
 - Fail fast!  Don't bother wasting time trying to start a service when we know the configuration is borked
 - Reduce Noise: No useful logs conf from a service that tries to start with a known-bad config
 - Surface lifecycle information: Log information about the image and the attempted start time. This makes it easier to parse the logs to figure out when a service is flagging or stable. It can be clunky to parse the AWS task event logs to get this information from 'outside' of the container, for example
 - the preflight config file can be a contract/manifest between dev devops: ideally, the service engineering team is running preflight in their development containers from the beginning.  That probably helps dev, since they are their own devops in their local environments.  Then moving to shared deployments that are managed by dev, the preflight config is a part of the image, streamling the process for dev to provide all the right config for a new service


example preflight cnofig:
```yaml
verbose: True
organization: "MyCompanyName"
team: "DevOps"
checked_environment_variables:
  - IS_AWS

```

```shell script
#!/usr/bin/env bash
#  CRITICAL:  set -e  to exit the script and prevent the service from being started if preflight exits 
#  with a non-zero code
set -e
preflight
echo "YOU SHOULD NOT SEE ME IF preflight FAILS"
/bath/to/service start
```

preflight is intended to check all of the lifecycle stuff that devops has to configure to get a service to run.  It should help us fail fast and surface all of the container configuration problems at once.  

The preflight configuration file defines the requirements. it cna be used as a sort of contract between the service engineers and devops.  It can be implemented in developer environments during their development cycle, then shared with devops for a clear and validated hand-off for easy deployment.
 
preflight is run inside a container in an entrypoint.sh script before the service start command. The entrypoint.sh script must be configured to exit on errors so a non-zero exit from preflight prevents the execution of the service start command.  This is important because if any of the preflight checks fail, we don't want to waste time and generate a bunch of expected but useless service errors. Instead, we prefer to fail quickly and generate a comprehensive list of all of the preflight checks that failed to reduce the troubleshooting cycles in devops.  Additionally, we insert distinctive tags in the preflight log entries based on the audience.  Preflight logs are specifically intended for the DevOps audience, so a string like '[MyCompanyName:DevOps]' in preflight log messages might become a de facto standard way of specifying the intended audience for a log message.

This preflight approach may also solve some challenges in quickly identifying steady state for ECS tasks. The way events are logged in AWS make it a bit tricky to parse.  These custom lifecycle messages from inside the container might make it easier.
 
 
Using the list of environment variables in the config key 'checked_environment_variables' preflight will  :
 - make sure each one has a value set
 - based on a standard envrionment variable name format, test host name resolutiojn
 - based on a standard envrionment variable name format, test tcp connections to dependencies
 - based on a standard envrionment variable name format, test client connection (check credentials)
 
## Mock Service
Build a container with a mock service that needs lots of things (config, database, other services with randomly generated endpoints) and make sure it complains loudly and obviously when it doesn't get what it needs  
 
## Environment variable Format
To check connetions, we need to have several pieces of data in a fairly compact format. To express a connection to a postgres10  database where I store hot pickles:
POSTGRES10_HOT_PICKLES_USERNAME=pat
POSTGRES10_HOT_PICKLES_PASSWORD=badpassword
POSTGRES10_HOT_PICKLES_ADDRESS=db.domain.com
POSTGRES10_HOT_PICKLES_PORT=5432
POSTGRES10_HOT_PICKLES_SSLMODE=ca-verify
POSTGRES10_HOT_PICKLES_SSLCERTPATH=/path/to/cert

Other common attributes include:  VERSION, DESCRIPTION, TOKEN, TIMEOUT


Important notes:
- underscore is reserved as a separator
- the first part ("POSTGRES10") is used to let preflight find the right client to test the connection
- the last part is the attribute in the host object: ('USERNAME', "PASSWORD", etc.)
- everything between the first and last is the host id ("HOT_PICKLES)

so to preflight, this gets converted to a map like:
id: HOT_PICKLES
client: POSTGRES10
username: pat
password: badpassword
address: db.domain.com
port: 5432
sslmode: ca-verify
sslcertpath: /path/to/cert




## Check tcp connections to host and port

Verify tcp connectivity to a list of host/port environment variable paris specified in the configuration.  Commonly the service engineers specify the environment variables they want to use for host and port information.  DevOps injects deployment-specific values  for portability.


## Validate credentials and set credential environment variables
This checks the last  n versions of a secret to try to log into a required resource, like a database.  If successful, it sets the environment variable(s) to the valid credentials and logs the key and sha256 hash of the value.

NOTE: this may require a plugin model at some point to handle authentication tests for various kinds or resources.


## Limitationss
An execution of preflight can only look at a single connection of each type and it has to explicitly match a client type. example

This would invloke the postgres client check if the other prio checks passed. but the service can't have a second postgres client
```text
POSTGRES_USERNAME=
POSTGRES_PASSWORD=
```

if you had this, there would have to explicitly be a client for each type of postgres db even though the checks  for credentials are the same.  This leave open the problem of client versions, too. this is a opportunity for improvement
```text
FRUITS_POSTGRES_USERNAME=
FRUITS_POSTGRES_PASSWORD=
VEGGIES_POSTGRES_USERNAME=
VEGGIES_POSTGRES_PASSWORD=
```

## TODO
- add org and team to al log messages
- add connection checks
- add a sample client credential test
- add log message to indicate task ID an task image name/version, image connectivity data (ip, etc)
- all the check functions are similar. maybe use an interface to make it easier to extend
- group related connection env vars by client type / connection identifier / connection field  ex: POSTGRES92_JUNKBIN_USERNAME.  there can be many separators in the middle. the first and last have to be  from a reserved set.

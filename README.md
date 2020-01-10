# Preflight

Check the container environment before starting the service based on the configured requirements. The intent is that the service engineers manage a preflight config file and the devops team uses preflight to make sure all of the requirements are met.

## Check environment variables
In normal mode, log the required environment variables and a sha256 hash of their variables to be able to securely verify configurations.
In debug mode, also securely log all of the other environment variables in case a needed variable was entered with a key typo.

## Check tcp connections to host and port

Verify tcp connectivity to a list of host/port environment variable paris specified in the configuration.  Commonly the service engineers specify the environment variables they want to use for host and port information.  DevOps injects deployment-specific values  for portability.


## Validate credentials and set credential environment variables
This checks the last  n versions of a secret to try to log into a required resource, like a database.  If successful, it sets the environment variable(s) to the valid credentials and logs the key and sha256 hash of the value.

NOTE: this may require a plugin model at some point to handle authentication tests for various kinds or resources.

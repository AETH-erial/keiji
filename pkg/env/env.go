package env

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const IMAGE_STORE = "IMAGE_STORE"
const WEB_ROOT = "WEB_ROOT"
const DOMAIN_NAME = "DOMAIN_NAME"
const HOST_PORT = "HOST_PORT"
const HOST_ADDR = "HOST_ADDR"
const SSL_MODE = "SSL_MODE"
const CHAIN = "CHAIN"
const KEY = "KEY"

var OPTION_VARS = map[string]string{
	IMAGE_STORE: "#the location for keiji to store the images uploaded (string)",
	WEB_ROOT:    "#the location to pull HTML and various web assets from. Only if using 'keiji -content fs' (string)",
	CHAIN:       "#the path to the SSL public key chain (string)",
	KEY:         "#the path to the SSL private key (string)",
}

var REQUIRED_VARS = map[string]string{
	HOST_PORT:   "#the port to run the server on (int)",
	HOST_ADDR:   "#the address for the server to listen on (string)",
	DOMAIN_NAME: "#the servers domain name, i.e. 'aetherial.dev', or 'localhost' (string)",
	SSL_MODE:    "#chose to use SSL or not (boolean)",
}

type EnvNotSet struct {
	NotSet string
}

func (e *EnvNotSet) Error() string {
	return fmt.Sprintf("Environment variable: '%s' was not set.", e.NotSet)
}

/*
Write out a blank .env configuration with the the required configuration (uncommented) and the
optional configuration (commented out)

	:param path: the path to write the template to
*/
func WriteTemplate(path string) {
	var out string
	out = out + "####### Required Configuration #######\n"
	for k, v := range REQUIRED_VARS {
		out = out + fmt.Sprintf("%s\n%s=\n", v, k)
	}
	out = out + "\n####### Optional Configuration #######\n"
	for k, v := range OPTION_VARS {
		out = out + fmt.Sprintf("# %s\n# %s=\n", v, k)
	}
	err := os.WriteFile(path, []byte(out), os.ModePerm)
	if err != nil {
		log.Fatal("Failed to write file: ", err)
	}

}

/*
verify all environment vars passed in are set

	:param vars: array of strings to verify
*/
func LoadAndVerifyEnv(path string, vars map[string]string) error {

	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	for i := range vars {
		if os.Getenv(vars[i]) == "" {
			return &EnvNotSet{NotSet: vars[i]}
		}
	}
	return nil
}

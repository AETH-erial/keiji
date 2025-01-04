package env

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/joho/godotenv"
)

const KEIJI_USERNAME = "KEIJI_USERNAME"
const KEIJI_PASSWORD = "KEIJI_PASSWORD"
const IMAGE_STORE = "IMAGE_STORE"
const WEB_ROOT = "WEB_ROOT"
const DOMAIN_NAME = "DOMAIN_NAME"
const HOST_PORT = "HOST_PORT"
const HOST_ADDR = "HOST_ADDR"
const USE_SSL = "USE_SSL"
const CHAIN = "CHAIN"
const KEY = "KEY"

var OPTION_VARS = map[string]string{
	IMAGE_STORE: "#the location for keiji to store the images uploaded (string)",
	WEB_ROOT:    "#the location to pull HTML and various web assets from. Only if using 'keiji -content fs' (string)",
	CHAIN:       "#the path to the SSL public key chain (string)",
	KEY:         "#the path to the SSL private key (string)",
}

var REQUIRED_VARS = map[string]string{
	HOST_PORT:      "#the port to run the server on (int)",
	HOST_ADDR:      "#the address for the server to listen on (string)",
	DOMAIN_NAME:    "#the servers domain name, i.e. 'aetherial.dev', or 'localhost' (string)",
	USE_SSL:        "#chose to use SSL or not (boolean)",
	KEIJI_USERNAME: "#the administrator username to login with",
	KEIJI_PASSWORD: "#the password for the administrator accounit",
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
func WriteTemplate(wtr io.Writer) error {

	outReqArr := make([]string, len(REQUIRED_VARS))
	outOptVar := make([]string, len(OPTION_VARS))
	i := 0
	for k := range REQUIRED_VARS {
		outReqArr[i] = k
		i++
	}
	i = 0
	for k := range OPTION_VARS {
		outOptVar[i] = k
		i++
	}
	sort.Strings(outReqArr)
	sort.Strings(outOptVar)

	var out string
	out = out + "####### Required Configuration #######\n"
	for i := range outReqArr {
		k := REQUIRED_VARS[outReqArr[i]]
		v := outReqArr[i]
		fmt.Println(k, v)
		out = out + fmt.Sprintf("%s\n%s=\n", v, k)
	}
	out = out + "\n####### Optional Configuration #######\n"
	for i := range outOptVar {
		out = out + fmt.Sprintf("# %s\n# %s=\n", OPTION_VARS[outOptVar[i]], outOptVar[i])
	}
	msg := []byte(out)
	_, err := io.Copy(wtr, bytes.NewBuffer(msg))
	if err != nil {
		return err
	}
	return nil
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

	for k := range vars {
		if os.Getenv(k) == "" {
			return &EnvNotSet{NotSet: k}
		}
	}
	return nil
}

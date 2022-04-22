package generator

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

// variableExists checks if a variable exists in a slice of environment variables
func variableExists(vars *[]LagoonEnvironmentVariable, name, value string) bool {
	exists := false
	for _, v := range *vars {
		if v.Name == name && v.Value == value {
			exists = true
		}
	}
	return exists
}

func strPtr(str string) *string {
	v := str
	return &v
}

// in Lagoon this bash is used to generated the hash, but the `echo "${ROUTE_DOMAIN}"` adds a new line at the end, so the generated
// sum has this in it, we need to replicate this here
// INGRESS_NAME="${ROUTE_DOMAIN:0:47}"
// INGRESS_NAME="${INGRESS_NAME%%.*}-$(echo "${ROUTE_DOMAIN}" | md5sum | cut -f 1 -d " " | cut -c 1-5)"
func getMD5HashWithNewLine(text string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s\n", text)))
	return hex.EncodeToString(hash[:])
}

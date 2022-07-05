package helpers

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// StrPtr .
func StrPtr(str string) *string {
	v := str
	return &v
}

// BoolPtr .
func BoolPtr(b bool) *bool {
	v := b
	return &v
}

// IntPTr
func IntPtr(i int) *int {
	return &i
}

// GetMD5HashWithNewLine .
// in Lagoon this bash is used to generated the hash, but the `echo "${ROUTE_DOMAIN}"` adds a new line at the end, so the generated
// sum has this in it, we need to replicate this here
// INGRESS_NAME="${ROUTE_DOMAIN:0:47}"
// INGRESS_NAME="${INGRESS_NAME%%.*}-$(echo "${ROUTE_DOMAIN}" | md5sum | cut -f 1 -d " " | cut -c 1-5)"
func GetMD5HashWithNewLine(text string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s\n", text)))
	return hex.EncodeToString(hash[:])
}

// GetSha256Hash gets the sha256 hash of a string.
func GetSha256Hash(text string) []byte {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s", text)))
	return hash[:]
}

// GetBase32EncodedLowercase gets a lowercase version of some data in base32 encoding.
// Lagoon does this currently, so replicating behaviour.
func GetBase32EncodedLowercase(data []byte) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(data)[:])
}

// GetEnv gets an environment variable
func GetEnv(key, fallback string, debug bool) string {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			if debug {
				fmt.Println(fmt.Sprintf("Using value from environment variable %s", key))
			}
			return value
		}
	}
	return fallback
}

// Contains checks if a string slice contains a specific string.
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// WriteTemplateFile writes the template to a file.
func WriteTemplateFile(templateOutputFile string, data []byte) {
	err := os.WriteFile(templateOutputFile, data, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
}

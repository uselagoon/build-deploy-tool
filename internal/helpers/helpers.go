package helpers

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base32"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
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

// Int32Ptr
func Int32Ptr(i int32) *int32 {
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

// GetEnvInt get environment variable as int, assume failed conversion is not existing
func GetEnvInt(key string, fallback int, debug bool) int {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			valueInt, e := strconv.Atoi(value)
			if e == nil {
				if debug {
					fmt.Println(fmt.Sprintf("Using value from environment variable %s", key))
				}
				return valueInt
			}
		}
	}
	return fallback
}

// EGetEnvInt get environment variable as int, return error if the value can't be converted
func EGetEnvInt(key string, fallback int, debug bool) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			valueInt, e := strconv.Atoi(value)
			if e != nil {
				return 0, fmt.Errorf("Unable to convert value %s to integer: %v", value, e)
			}
			if debug {
				fmt.Println(fmt.Sprintf("Using value from environment variable %s", key))
			}
			return valueInt, nil
		}
	}
	return fallback, nil
}

// GetEnvBool get environment variable as bool
// accepts fallback values 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False
// anything else is false.
func GetEnvBool(key string, fallback bool, debug bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			rVal, _ := strconv.ParseBool(value)
			if debug {
				fmt.Println(fmt.Sprintf("Using value from environment variable %s", key))
			}
			return rVal
		}
	}
	return fallback
}

// EGetEnvBool get environment variable as bool, return error if the value can't be converted
// accepts fallback values 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False
// anything else is false.
func EGetEnvBool(key string, fallback bool, debug bool) (bool, error) {
	if value, ok := os.LookupEnv(key); ok {
		if value != "" {
			rVal, e := strconv.ParseBool(value)
			if e != nil {
				return false, fmt.Errorf("Unable to convert value %s to bool: %v", value, e)
			}
			if debug {
				fmt.Println(fmt.Sprintf("Using value from environment variable %s", key))
			}
			return rVal, nil
		}
	}
	return fallback, nil
}

func StrToBool(value string) bool {
	rVal, e := strconv.ParseBool(value)
	if e != nil {
		return false
	}
	return rVal
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

type EnvironmentVariable struct {
	Name  string
	Value string
}

// Try and get the namespace name from the serviceaccount location if it exists
func GetNamespace(namespace, filename string) (string, error) {
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		nsb, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		namespace = strings.Trim(string(nsb), "\n ")
	}
	return namespace, nil
}

func UnsetEnvVars(localVars []EnvironmentVariable) {
	varNames := []string{
		"MONITORING_ALERTCONTACT",
		"MONITORING_STATUSPAGEID",
		"PROJECT",
		"ENVIRONMENT",
		"BRANCH",
		"PR_NUMBER",
		"PR_HEAD_BRANCH",
		"PR_BASE_BRANCH",
		"ENVIRONMENT_TYPE",
		"BUILD_TYPE",
		"ACTIVE_ENVIRONMENT",
		"STANDBY_ENVIRONMENT",
		"LAGOON_FASTLY_NOCACHE_SERVICE_ID",
		"LAGOON_PROJECT_VARIABLES",
		"LAGOON_ENVIRONMENT_VARIABLES",
		"LAGOON_VERSION",
		"ROUTE_FASTLY_SERVICE_ID",
		"FASTLY_API_SECRET_PREFIX",
		"LAGOON_FEATURE_BACKUP_DEV_SCHEDULE",
		"LAGOON_FEATURE_BACKUP_PR_SCHEDULE",
		"LAGOON_GIT_BRANCH",
		"DEFAULT_BACKUP_SCHEDULE",
		"LAGOON_FEATURE_BACKUP_DEV_SCHEDULE",
		"LAGOON_FEATURE_BACKUP_PR_SCHEDULE",
		"LAGOON_FEATURE_FLAG_DEFAULT_INGRESS_CLASS",
	}
	for _, varName := range varNames {
		os.Unsetenv(varName)
	}
	for _, varName := range localVars {
		os.Unsetenv(varName.Name)
	}
}

func CheckLabelLength(labels map[string]string) error {
	for k, v := range labels {
		if len(v) > 63 {
			return fmt.Errorf("label: '%s', value: '%s' is longer than 63 characters", k, v)
		}
	}
	return nil
}

func DeepCopy(src, dist interface{}) (err error) {
	buf := bytes.Buffer{}
	if err = gob.NewEncoder(&buf).Encode(src); err != nil {
		return
	}
	return gob.NewDecoder(&buf).Decode(dist)
}

func AppendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

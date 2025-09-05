package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/distribution/reference"

	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"k8s.io/apimachinery/pkg/api/resource"
)

// checks the provided environment variables looking for feature flag based variables
func CheckFeatureFlag(key string, envVariables []lagoon.EnvironmentVariable, debug bool) string {
	// check for force value
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_FORCE_%s", key)); ok {
		if debug {
			fmt.Printf("Using forced flag value from build variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_FORCE_%s", key))
		}
		return value
	}
	// check lagoon environment variables
	for _, lVar := range envVariables {
		if strings.Contains(lVar.Name, fmt.Sprintf("LAGOON_FEATURE_FLAG_%s", key)) {
			if debug {
				fmt.Printf("Using flag value from Lagoon environment variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_%s", key))
			}
			return lVar.Value
		}
	}
	// return default
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_DEFAULT_%s", key)); ok {
		if debug {
			fmt.Printf("Using default flag value from build variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_DEFAULT_%s", key))
		}
		return value
	}
	// otherwise nothing
	return ""
}

func CheckAdminFeatureFlag(key string, debug bool) string {
	if value, ok := os.LookupEnv(fmt.Sprintf("ADMIN_LAGOON_FEATURE_FLAG_%s", key)); ok {
		if debug {
			fmt.Printf("Using admin feature flag value from build variable %s\n", fmt.Sprintf("ADMIN_LAGOON_FEATURE_FLAG_%s", key))
		}
		return value
	}
	return ""
}

func ValidateResourceQuantity(s string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New(fmt.Sprint(x))
			}
		}
	}()
	resource.MustParse(s)
	return nil
}

func ValidateResourceSize(size string) (int64, error) {
	volQ, err := resource.ParseQuantity(size)
	if err != nil {
		return 0, err
	}
	volS, _ := volQ.AsInt64()
	return volS, nil
}

// ContainsRegistry checks if a string slice contains a specific string regex match.
func ContainsRegistry(regex []ContainerRegistry, match string) bool {
	for _, v := range regex {
		m, _ := regexp.MatchString(v.URL, match)
		if m {
			return true
		}
	}
	return false
}

func checkDuplicateCronjobs(cronjobs []lagoon.Cronjob) error {
	var unique []lagoon.Cronjob
	var duplicates []lagoon.Cronjob
	for _, v := range cronjobs {
		skip := false
		for _, u := range unique {
			if v.Name == u.Name {
				skip = true
				duplicates = append(duplicates, v)
				break
			}
		}
		if !skip {
			unique = append(unique, v)
		}
	}
	var uniqueDuplicates []lagoon.Cronjob
	for _, d := range duplicates {
		for _, u := range unique {
			if d.Name == u.Name {
				uniqueDuplicates = append(uniqueDuplicates, u)
			}
		}
	}
	// join the two together
	result := append(duplicates, uniqueDuplicates...)
	if result != nil {
		b, _ := json.Marshal(result)
		return fmt.Errorf("duplicate named cronjobs detected: %v", string(b))
	}
	return nil
}

// getDBaasEnvironment will check the dbaas provider to see if an environment exists or not
func getDBaasEnvironment(
	buildValues *BuildValues,
	dbaasEnvironment *string,
	lagoonOverrideName,
	lagoonType string,
) (bool, error) {
	if buildValues.DBaaSEnvironmentTypeOverrides != nil {
		dbaasEnvironmentTypeSplit := strings.Split(buildValues.DBaaSEnvironmentTypeOverrides.Value, ",")
		for _, sType := range dbaasEnvironmentTypeSplit {
			sTypeSplit := strings.Split(sType, ":")
			if sTypeSplit[0] == lagoonOverrideName {
				*dbaasEnvironment = sTypeSplit[1]
			}
		}
	}
	exists, err := buildValues.DBaaSClient.CheckProvider(buildValues.DBaaSOperatorEndpoint, lagoonType, *dbaasEnvironment)
	if err != nil {
		return exists, fmt.Errorf("there was an error checking DBaaS endpoint %s: %v", buildValues.DBaaSOperatorEndpoint, err)
	}
	return exists, nil
}

var exp = regexp.MustCompile(`(\\*)\$\{(.+?)(?:(\:\-)(.*?))?\}`)

func determineRefreshImage(serviceName, imageName string, envVars []lagoon.EnvironmentVariable) (string, []error) {
	errs := []error{}
	parsed := exp.ReplaceAllStringFunc(string(imageName), func(match string) string {
		tagvalue := ""
		re := regexp.MustCompile(`\${?(\w+)?(?::-(\w+))?}?`)
		matches := re.FindStringSubmatch(match)
		if len(matches) > 0 {
			tv := ""
			envVarKey := matches[1]
			defaultVal := matches[2] //This could be empty
			for _, v := range envVars {
				if v.Name == envVarKey {
					tv = v.Value
				}
			}
			if tv == "" {
				if defaultVal != "" {
					tagvalue = defaultVal
				} else {
					errs = append(errs, fmt.Errorf("the 'lagoon.base.image' label defined on service %s in the docker-compose file is invalid ('%s') - no matching variable or fallback found to replace requested variable %s", serviceName, imageName, envVarKey))
				}
			} else {
				tagvalue = tv
			}
		}
		return tagvalue
	})
	if parsed == imageName {
		if !reference.ReferenceRegexp.MatchString(parsed) {
			if strings.Contains(parsed, "$") {
				errs = append(errs, fmt.Errorf("the 'lagoon.base.image' label defined on service %s in the docker-compose file is invalid ('%s') - variables are defined incorrectly, must contain curly brackets (example: '${VARIABLE}')", serviceName, imageName))
			} else {
				errs = append(errs, fmt.Errorf("the 'lagoon.base.image' label defined on service %s in the docker-compose file is invalid ('%s') - please ensure it conforms to the structure `[REGISTRY_HOST[:REGISTRY_PORT]/]REPOSITORY[:TAG|@DIGEST]`", serviceName, imageName))
			}
		}
	}
	return parsed, errs
}

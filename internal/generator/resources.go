package generator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func getResourcesFromAPIEnvVar(
	envVars []lagoon.EnvironmentVariable,
	debug bool,
) (*map[string]ResourceWorkloads, error) {
	resWorkloads := &map[string]ResourceWorkloads{}
	// TODO: this is still to be determined how the data will be consumed from the API, it may eventually come from
	// a configmap or some other means, or a combination of configmap and envvar merging
	// for now, consume from featureflag var
	resourceWorkloadsJSON := CheckFeatureFlag("WORKLOAD_RESOURCES", envVars, debug)
	// only from envvar from api, not feature flagable
	// resourceWorkloadsJSONvar, _ := lagoon.GetLagoonVariable("LAGOON_WORKLOAD_RESOURCES", []string{"build", "global"}, envVars)
	// if resourceWorkloadsJSONvar != nil {
	// 	resourceWorkloadsJSON = resourceWorkloadsJSONvar.Value
	// }
	if resourceWorkloadsJSON != "" {
		if debug {
			fmt.Println("Collecting resource workloads from WORKLOAD_RESOURCES variable")
		}
		// if the routesJSON is populated, then attempt to decode and unmarshal it
		rawJSONStr, err := base64.StdEncoding.DecodeString(resourceWorkloadsJSON)
		if err != nil {
			return nil, fmt.Errorf("couldn't decode resource workloads from Lagoon API, is it actually base64 encoded?: %v", err)
		}
		rawJSON := []byte(rawJSONStr)
		err = json.Unmarshal(rawJSON, resWorkloads)
		if err != nil {
			return nil, fmt.Errorf("couldn't unmarshal resource workloads from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
		}
	}
	return resWorkloads, nil
}

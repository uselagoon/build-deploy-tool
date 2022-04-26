package lagoon

import (
	"fmt"
	"strconv"
	"strings"
)

// GenerateFastlyConfiguration generates the fastly configuration for a specific route from Lagoon variables.
func GenerateFastlyConfiguration(noCacheServiceID, serviceID, route string, variables []EnvironmentVariable) (Fastly, error) {
	f := Fastly{}
	if serviceID == "" {
		if noCacheServiceID != "" {
			f.ServiceID = noCacheServiceID
			f.Watch = true
		}
	}
	// check lagoon api variables for `LAGOON_FASTLY_SERVICE_ID`
	// this is supported as `SERVICE_ID:WATCH_STATUS:SECRET_NAME(optional)` eg: "fa23rsdgsdgas:false", "fa23rsdgsdgas:true" or "fa23rsdgsdgas:true:examplecom"
	// this will apply to ALL ingresses if one is not specifically defined in the `LAGOON_FASTLY_SERVICE_IDS` environment variable override
	// see section `FASTLY SERVICE ID PER INGRESS OVERRIDE` in `build-deploy-docker-compose.sh` for info on `LAGOON_FASTLY_SERVICE_IDS`
	lfsID, err := GetLagoonVariable("LAGOON_FASTLY_SERVICE_ID", variables)
	if err == nil {
		lfsIDSplit := strings.Split(lfsID.Value, ":")
		if len(lfsIDSplit) == 1 {
			return f, fmt.Errorf("no watch status was provided, only the service id")
		}
		watch, err := strconv.ParseBool(lfsIDSplit[1])
		if err != nil {
			return f, fmt.Errorf("the provided value %s is not a valid boolean", lfsIDSplit[1])
		}
		f.ServiceID = lfsIDSplit[0]
		f.Watch = watch
		if len(lfsIDSplit) == 3 {
			// the optional secret has been defined
			f.APISecretName = lfsIDSplit[2]
		}
	}
	// check the `LAGOON_FASTLY_SERVICE_IDS` to see if we have a domain specific override
	// this is useful if all domains are using the nocache service, but you have a specific domain that should use a different service
	// and you haven't defined it in the lagoon.yml file
	// see section `FASTLY SERVICE ID PER INGRESS OVERRIDE` in `build-deploy-docker-compose.sh` for info on `LAGOON_FASTLY_SERVICE_IDS`
	lfsIDs, err := GetLagoonVariable("LAGOON_FASTLY_SERVICE_IDS", variables)
	if err == nil {
		lfsIDsSplit := strings.Split(lfsIDs.Value, ",")
		for _, lfs := range lfsIDsSplit {
			lfsIDSplit := strings.Split(lfs, ":")
			if lfsIDSplit[0] == route {
				if len(lfsIDSplit) == 2 {
					return f, fmt.Errorf("no watch status was provided, only the route and service id")
				}
				watch, err := strconv.ParseBool(lfsIDSplit[2])
				if err != nil {
					return f, fmt.Errorf("the provided value %s is not a valid boolean", lfsIDSplit[2])
				}
				f.ServiceID = lfsIDSplit[1]
				f.Watch = watch
				// unset the apisecret name if this point is reached
				// this is because this particular ingress may not have one defined
				// it will get checked next
				f.APISecretName = ""
				if len(lfsIDSplit) == 4 {
					// the optional secret has been defined
					f.APISecretName = lfsIDSplit[3]
				}
			}
		}
	}
	return f, nil
}

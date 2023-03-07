package generator

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alessio/shellescape"
	composetypes "github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
)

// this is a map that maps old service types to their new service types
var oldServiceMap = map[string]string{
	"mariadb-shared":        "mariadb-dbaas",
	"postgres-shared":       "postgres-dbaas",
	"mongo-shared":          "mongodb-dbaas",
	"python-ckandatapusher": "python",
	"mongo":                 "mongodb",
}

// these are lagoon types that support autogenerated routes
var supportedAutogeneratedTypes = []string{
	// "kibana", //@TODO: don't even need this anymore?
	"basic",
	"basic-persistent",
	"basic-single",
	"node",
	"node-persistent",
	"nginx",
	"nginx-php",
	"nginx-php-persistent",
	"varnish",
	"varnish-persistent",
	"python-persistent",
	"python",
}

// these service types don't have images
var ignoredImageTypes = []string{
	"mariadb-dbaas",
	"postgres-dbaas",
	"mongodb-dbaas",
}

// these are lagoon types that support autogenerated routes
var supportedDBTypes = []string{
	"mariadb",
	"mariadb-dbaas",
	"postgres",
	"postgres-dbaas",
	"mongodb",
	"mongodb-dbaas",
}

// these are lagoon types that come with resources requiring backups
var typesWithBackups = []string{
	"basic-persistent",
	"basic-single",
	"node-persistent",
	"nginx-php-persistent",
	"python-persistent",
	"varnish-persistent",
	"redis-persistent",
	"solr",
	"elasticsearch",
	"opensearch",
	"rabbitmq",
	"mongodb-dbaas",
	"mariadb-dbaas",
	"postgres-dbaas",
	"mariadb-single",
	"postgres-single",
	"mongodb-single",
}

// this is commented out as this is not enforced currently, but some enforcement should be implemented
// var (
// 	maxServices int = 10
// )

// generateServicesFromDockerCompose unmarshals the docker-compose file and processes the services using composeToServiceValues
func generateServicesFromDockerCompose(
	buildValues *BuildValues,
	ignoreNonStringKeyErrors, ignoreMissingEnvFiles, debug bool,
) error {
	// take lagoon envvars and create new map for being unmarshalled against the docker-compose file
	composeVars := make(map[string]string)
	for _, envvar := range buildValues.EnvironmentVariables {
		composeVars[envvar.Name] = envvar.Value
	}

	// create the services map
	buildValues.Services = []ServiceValues{}

	// unmarshal the docker-compose.yml file
	lCompose, lComposeOrder, lComposeVolumes, err := lagoon.UnmarshaDockerComposeYAML(
		buildValues.LagoonYAML.DockerComposeYAML,
		ignoreNonStringKeyErrors,
		ignoreMissingEnvFiles,
		composeVars,
	)
	if err != nil {
		return err
	}

	// convert docker-compose volumes to buildvolumes,
	err = convertVolumes(buildValues, lCompose, lComposeVolumes)
	if err != nil {
		return err
	}

	// convert docker-compose services to servicevalues,
	// range over the original order of the docker-compose file when setting services
	for _, service := range lComposeOrder {
		for _, composeServiceValues := range lCompose.Services {
			if service.Name == composeServiceValues.Name {
				cService, err := composeToServiceValues(buildValues, composeServiceValues.Name, composeServiceValues, debug)
				if err != nil {
					return err
				}
				if cService != nil {
					if cService.BackupsEnabled {
						buildValues.BackupsEnabled = true
					}
					buildValues.Services = append(buildValues.Services, *cService)
				}
				// to prevent too many services from being provisioned, some sort of limit should probably be imposed
				// this is commented out as this is not enforced currently, but some enforcement should be implemented
				// if len(buildValues.Services) > maxServices {
				// 	return fmt.Errorf("unable to provision more than %d services for this environment, if you need more please contact your lagoon administrator", maxServices)
				// }
			}
		}
	}
	return nil
}

// composeToServiceValues is the primary function used to pre-seed how templates are created
// it reads the docker-compose file and converts each service into a ServiceValues struct
// this is the "known state" of that service, and all subsequent steps to create templates will use this data unmodified
func composeToServiceValues(
	buildValues *BuildValues,
	composeService string,
	composeServiceValues composetypes.ServiceConfig,
	debug bool,
) (*ServiceValues, error) {
	lagoonType := ""
	// if there are no labels, then this is probably not going to end up in Lagoon
	// the lagoonType check will skip to the end and return an empty service definition
	if composeServiceValues.Labels != nil {
		lagoonType = lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.type")
	}
	if lagoonType == "" {
		return nil, fmt.Errorf(
			"no lagoon.type has been set for service %s. If a Lagoon service is not required, please set the lagoon.type to 'none' for this service in docker-compose.yaml. See the Lagoon documentation for supported service types",
			composeService,
		)
	} else {
		// if the lagoontype is populated, even none is valid as there may be a servicetype override in an environment variable
		autogenEnabled := true
		autogenTLSAcmeEnabled := true
		autogeRequestVerification := false
		// check if autogenerated routes are disabled
		if buildValues.LagoonYAML.Routes.Autogenerate.Enabled != nil {
			if !*buildValues.LagoonYAML.Routes.Autogenerate.Enabled {
				autogenEnabled = false
			}
		}
		// check if pullrequests autogenerated routes are disabled
		if buildValues.BuildType == "pullrequest" && buildValues.LagoonYAML.Routes.Autogenerate.AllowPullRequests != nil {
			if !*buildValues.LagoonYAML.Routes.Autogenerate.AllowPullRequests {
				autogenEnabled = false
			} else {
				autogenEnabled = true
			}
		}
		// check if this environment has autogenerated routes disabled
		if buildValues.LagoonYAML.Environments[buildValues.Branch].AutogenerateRoutes != nil {
			if !*buildValues.LagoonYAML.Environments[buildValues.Branch].AutogenerateRoutes {
				autogenEnabled = false
			} else {
				autogenEnabled = true
			}
		}
		// check if autogenerated routes tls-acme disabled
		if buildValues.LagoonYAML.Routes.Autogenerate.TLSAcme != nil {
			if !*buildValues.LagoonYAML.Routes.Autogenerate.TLSAcme {
				autogenTLSAcmeEnabled = false
			}
		}
		// check if autogenerated routes request verification disabled
		if buildValues.LagoonYAML.Routes.Autogenerate.RequestVerification != nil {
			if !*buildValues.LagoonYAML.Routes.Autogenerate.RequestVerification {
				autogeRequestVerification = false
			} else {
				autogeRequestVerification = true
			}
		}
		// check lagoon yaml for an override for this service
		if value, ok := buildValues.LagoonYAML.Environments[buildValues.Environment].Types[composeService]; ok {
			lagoonType = value
		}
		// check if the service has a specific override
		serviceAutogenerated := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.autogeneratedroute")
		if serviceAutogenerated != "" {
			if reflect.TypeOf(serviceAutogenerated).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(serviceAutogenerated)
				if err == nil {
					autogenEnabled = vBool
				}
			}
		}
		// check if the service has a tls-acme specific override
		serviceAutogeneratedTLSAcme := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.autogeneratedroute.tls-acme")
		if serviceAutogeneratedTLSAcme != "" {
			if reflect.TypeOf(serviceAutogeneratedTLSAcme).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(serviceAutogeneratedTLSAcme)
				if err == nil {
					autogenTLSAcmeEnabled = vBool
				}
			}
		}
		// check if the service has a deployment servicetype override
		// @TODO: this was previously used to detect which image to use for a linked service, but the logic for that has changed now
		// this isn't required anymore. leaving the check here for now but `serviceDeploymentServiceType` is currently unused
		// REF: https://github.com/uselagoon/build-deploy-tool/blob/997483b59ac7c055a07f04f579b91ccd7e4fb4a2/legacy/build-deploy-docker-compose.sh#L439-L444
		serviceDeploymentServiceType := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.deployment.servicetype")
		if serviceDeploymentServiceType == "" {
			serviceDeploymentServiceType = composeService
		}

		// if there is a `lagoon.name` label on this service, this should be used as an override name
		lagoonOverrideName := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.name")
		if lagoonOverrideName != "" {
			// if there is an override name, check all other services already existing
			for _, service := range buildValues.Services {
				// if there is an existing service with this same override name, then disable autogenerated routes
				// for this service
				if service.OverrideName == lagoonOverrideName {
					autogenEnabled = false
				}
			}
		} else {
			// otherwise just set the override name to be the service name
			lagoonOverrideName = composeService
		}

		// if there are overrides defined in the lagoon API `LAGOON_SERVICE_TYPES`
		// handle those here
		if buildValues.ServiceTypeOverrides != nil {
			serviceTypesSplit := strings.Split(buildValues.ServiceTypeOverrides.Value, ",")
			for _, sType := range serviceTypesSplit {
				sTypeSplit := strings.Split(sType, ":")
				if sTypeSplit[0] == lagoonOverrideName {
					lagoonType = sTypeSplit[1]
				}
			}
		}

		// convert old service types to new service types from the old service map
		// this allows for adding additional values to the oldServiceMap that we can force to be anything else
		if val, ok := oldServiceMap[lagoonType]; ok {
			lagoonType = val
		}

		// if there are no overrides, and the type is none, then abort here, no need to proceed calculating the type
		if lagoonType == "none" {
			return nil, nil
		}
		// anything after this point is where heavy processing is done as the service type has now been determined by this stage

		// handle dbaas operator checks here
		dbaasEnvironment := buildValues.EnvironmentType
		svcIsDBaaS := false
		svcIsSingle := false
		if helpers.Contains(supportedDBTypes, lagoonType) {
			// strip the dbaas off the supplied type for checking against providers, it gets added again later
			lagoonType = strings.Split(lagoonType, "-dbaas")[0]
			err := buildValues.DBaaSClient.CheckHealth(buildValues.DBaaSOperatorEndpoint)
			if err != nil {
				// @TODO eventually this error should be handled and fail a build, with a flag to override https://github.com/uselagoon/build-deploy-tool/issues/56
				// if !buildValues.DBaaSFallbackSingle {
				// 	return nil, fmt.Errorf("unable to check the DBaaS endpoint %s: %v", buildValues.DBaaSOperatorEndpoint, err)
				// }
				if debug {
					fmt.Printf("Unable to check the DBaaS endpoint %s, falling back to %s-single: %v\n", buildValues.DBaaSOperatorEndpoint, lagoonType, err)
				}
				// normally we would fall back to doing a cluster capability check, this is phased out in the build tool, it isn't reliable
				// and noone should be doing checks that way any more
				// the old bash check is the following
				// elif [[ "${CAPABILITIES[@]}" =~ "mariadb.amazee.io/v1/MariaDBConsumer" ]] && ! checkDBaaSHealth ; then
				lagoonType = fmt.Sprintf("%s-single", lagoonType)
			} else {
				// if there is a `lagoon.%s-dbaas.environment` label on this service, this should be used as an the environment type for the dbaas
				dbaasLabelOverride := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, fmt.Sprintf("lagoon.%s-dbaas.environment", lagoonType))
				if dbaasLabelOverride != "" {
					dbaasEnvironment = dbaasLabelOverride
				}

				// @TODO: maybe phase this out?
				// if value, ok := buildValues.LagoonYAML.Environments[buildValues.Environment].Overrides[composeService][mariadb][mariadb-dbaas].Environment; ok {
				// this isn't documented in the lagoon.yml, and it looks like a failover from days past.
				// 	lagoonType = value
				// }

				// if there are overrides defined in the lagoon API `LAGOON_DBAAS_ENVIRONMENT_TYPES`
				// handle those here
				exists, err := getDBaasEnvironment(buildValues, &dbaasEnvironment, lagoonOverrideName, lagoonType)
				if err != nil {
					// @TODO eventually this error should be handled and fail a build, with a flag to override https://github.com/uselagoon/build-deploy-tool/issues/56
					// if !buildValues.DBaaSFallbackSingle {
					// 	return nil, err
					// }
					if debug {
						fmt.Printf(
							"There was an error checking DBaaS endpoint %s, falling back to %s-single: %v\n",
							buildValues.DBaaSOperatorEndpoint, lagoonType, err,
						)
					}
				}

				// if the requested dbaas environment exists, then set the type to be the requested type with `-dbaas`
				if exists {
					lagoonType = fmt.Sprintf("%s-dbaas", lagoonType)
					svcIsDBaaS = true
				} else {
					// otherwise fallback to -single (if DBaaSFallbackSingle is enabled, otherwise it will error out prior)
					lagoonType = fmt.Sprintf("%s-single", lagoonType)
					svcIsSingle = true
				}
			}
		}

		// check if the service has any persistent labels, this is the path that the volume will be mounted to
		servicePersistentPath := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.persistent")
		if servicePersistentPath == "" {
			// if there is no persistent path, check if the service type has a default path
			if val, ok := servicetypes.ServiceTypes[lagoonType]; ok {
				servicePersistentPath = val.Volumes.PersistentVolumePath
				// check if the service type provides or consumes a default persistent volume
				if (val.ProvidesPersistentVolume || val.ConsumesPersistentVolume) && servicePersistentPath == "" {
					return nil, fmt.Errorf("label lagoon.persistent not defined for service %s, no valid mount path was found", composeService)
				}
			}
		}
		servicePersistentName := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.persistent.name")
		if servicePersistentName == "" && servicePersistentPath != "" {
			// if there is a persistent path defined, then set the persistent name to be the compose service if no persistent name is provided
			// persistent name is used by joined services like nginx/php or cli or worker pods to mount another service volume
			servicePersistentName = lagoonOverrideName
		}
		servicePersistentSize := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.persistent.size")
		if servicePersistentSize == "" {
			// if there is no persistent size, check if the service type has a default size allocated
			if val, ok := servicetypes.ServiceTypes[lagoonType]; ok {
				servicePersistentSize = val.Volumes.PersistentVolumeSize
				// check if the service type provides persistent volume, and that a size was detected
				if val.ProvidesPersistentVolume && servicePersistentSize == "" {
					return nil, fmt.Errorf("label lagoon.persistent.size not defined for service %s, no valid size was found", composeService)
				}
			}
		}
		if servicePersistentSize != "" {
			// check the provided size is a valid resource size for kubernetes
			_, err := ValidateResourceSize(servicePersistentSize)
			if err != nil {
				return nil, fmt.Errorf("provided persistent volume size for %s is not valid: %v", servicePersistentName, err)
			}
		}

		// if any `lagoon.base.image` labels are set, we note them for docker pulling
		// this allows us to refresh the docker-host's cache in cases where an image
		// may have an update without a change in tag (i.e. "latest" tagged images)
		baseimage := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.base.image")
		if baseimage != "" {
			baseImageWithTag, errs := determineRefreshImage(composeService, baseimage, buildValues.EnvironmentVariables)
			if len(errs) > 0 {
				for idx, err := range errs {
					if idx+1 == len(errs) {
						return nil, err
					} else {
						fmt.Println(err)
					}
				}
			}
			buildValues.ForcePullImages = append(buildValues.ForcePullImages, baseImageWithTag)
		}

		// calculate if this service needs any additional volumes attached from the calculated build volumes
		// additional volumes can only be attached to certain
		serviceVolumes, err := calculateServiceVolumes(buildValues, lagoonType, servicePersistentName, composeServiceValues.Labels)
		if err != nil {
			return nil, err
		}

		// start spot instance handling
		useSpot := false
		forceSpot := false
		cronjobUseSpot := false
		cronjobForceSpot := false
		spotTypes := ""
		cronjobSpotTypes := ""
		spotReplicas := int32(0)

		// these services can support multiple replicas in production
		// @TODO this should probably be an admin only feature flag though
		prodSpotReplicaTypes := CheckAdminFeatureFlag("SPOT_TYPE_REPLICAS_PRODUCTION", debug)
		if prodSpotReplicaTypes == "" {
			prodSpotReplicaTypes = "nginx,nginx-persistent,nginx-php,nginx-php-persistent"
		}
		devSpotReplicaTypes := CheckAdminFeatureFlag("SPOT_TYPE_REPLICAS_DEVELOPMENT", debug)
		if devSpotReplicaTypes == "" {
			devSpotReplicaTypes = ""
		}

		productionSpot := CheckFeatureFlag("SPOT_INSTANCE_PRODUCTION", buildValues.EnvironmentVariables, debug)
		developmentSpot := CheckFeatureFlag("SPOT_INSTANCE_DEVELOPMENT", buildValues.EnvironmentVariables, debug)
		if productionSpot == "enabled" && buildValues.EnvironmentType == "production" {
			spotTypes = CheckFeatureFlag("SPOT_INSTANCE_PRODUCTION_TYPES", buildValues.EnvironmentVariables, debug)
			cronjobSpotTypes = CheckFeatureFlag("SPOT_INSTANCE_PRODUCTION_CRONJOB_TYPES", buildValues.EnvironmentVariables, debug)
		}
		if developmentSpot == "enabled" && buildValues.EnvironmentType == "development" {
			spotTypes = CheckFeatureFlag("SPOT_INSTANCE_DEVELOPMENT_TYPES", buildValues.EnvironmentVariables, debug)
			cronjobSpotTypes = CheckFeatureFlag("SPOT_INSTANCE_DEVELOPMENT_CRONJOB_TYPES", buildValues.EnvironmentVariables, debug)
		}
		// check if the provided spot instance types against the current lagoonType
		for _, t := range strings.Split(spotTypes, ",") {
			if t != "" {
				tt := strings.Split(t, ":")
				if tt[0] == lagoonType {
					useSpot = true
					// check if the length of the split is more than one indicating that `force` is provided
					if len(tt) > 1 && tt[1] == "force" {
						forceSpot = true
					}
				}
			}
		}
		// check if the provided cronjob spot instance types against the current lagoonType
		for _, t := range strings.Split(cronjobSpotTypes, ",") {
			if t != "" {
				tt := strings.Split(t, ":")
				if tt[0] == lagoonType {
					cronjobUseSpot = true
					// check if the length of the split is more than one indicating that `force` is provided
					if len(tt) > 1 && tt[1] == "force" {
						cronjobForceSpot = true
					}
				}
			}
		}
		// check if the this service is production and can support 2 replicas on spot
		for _, t := range strings.Split(prodSpotReplicaTypes, ",") {
			if t != "" {
				if t == lagoonType && buildValues.EnvironmentType == "production" && useSpot {
					spotReplicas = 2
				}
			}
		}
		for _, t := range strings.Split(devSpotReplicaTypes, ",") {
			if t != "" {
				if t == lagoonType && buildValues.EnvironmentType == "development" && useSpot {
					spotReplicas = 2
				}
			}
		}
		// end spot instance handling

		// work out cronjobs for this service
		inpodcronjobs := []lagoon.Cronjob{}
		nativecronjobs := []lagoon.Cronjob{}
		// check if there are any duplicate named cronjobs
		if err := checkDuplicateCronjobs(buildValues.LagoonYAML.Environments[buildValues.Branch].Cronjobs); err != nil {
			return nil, err
		}
		if !buildValues.CronjobsDisabled {
			for idx, cronjob := range buildValues.LagoonYAML.Environments[buildValues.Branch].Cronjobs {
				// if this cronjob is meant for this service, add it
				if cronjob.Service == composeService {
					var err error
					inpod, err := helpers.IsInPodCronjob(cronjob.Schedule)
					if err != nil {
						return nil, fmt.Errorf("unable to validate crontab for cronjob %s: %v", cronjob.Name, err)
					}
					cronjob.Schedule, err = helpers.ConvertCrontab(buildValues.Namespace, cronjob.Schedule)
					if err != nil {
						return nil, fmt.Errorf("unable to convert crontab for cronjob %s: %v", cronjob.Name, err)
					}
					// handle setting the cronjob timeout here
					// can't be greater than 24hrs and must match go time duration https://pkg.go.dev/time#ParseDuration
					if cronjob.Timeout != "" {
						cronjobTimeout, err := time.ParseDuration(cronjob.Timeout)
						if err != nil {
							return nil, fmt.Errorf("unable to convert timeout for cronjob %s: %v", cronjob.Name, err)
						}
						// max cronjob timeout is 24 hours
						if cronjobTimeout > time.Duration(24*time.Hour) {
							return nil, fmt.Errorf("timeout for cronjob %s cannot be longer than 24 hours", cronjob.Name)
						}
					} else {
						// default cronjob timeout is 4h
						cronjob.Timeout = "4h"
					}
					// if the cronjob is inpod, or the cronjob has an inpod flag override
					if inpod || (cronjob.InPod != nil && *cronjob.InPod) {
						cmd := cronjob.Command
						// Lagoon enforces that only a single instance of a cronjob can run at any one time.
						// https://man7.org/linux/man-pages/man1/flock.1.html
						// https://www.gnu.org/savannah-checkouts/gnu/bash/manual/bash.html#Shell-Parameter-Expansion
						sha := sha256.New()
						sha.Write([]byte(fmt.Sprintf("%d %s", idx, cmd)))
						cmdSha := sha.Sum(nil)
						cronjob.Command = fmt.Sprintf("flock -n /tmp/cron.lock.%x -c %s", cmdSha, shellescape.Quote(cmd))
						inpodcronjobs = append(inpodcronjobs, cronjob)
					} else {
						// make the cronjob name kubernetes compliant
						cronjob.Name = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(fmt.Sprintf("cronjob-%s-%s", lagoonOverrideName, strings.ToLower(cronjob.Name)), "-")
						if len(cronjob.Name) > 52 {
							// if the cronjob name is longer than 52 characters
							// truncate it and add a hash of the name to it
							cronjob.Name = fmt.Sprintf("%s-%s", cronjob.Name[:45], helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(cronjob.Name))[:6])
						}
						nativecronjobs = append(nativecronjobs, cronjob)
					}
				}
			}
		}

		// check if this service is one that supports autogenerated routes
		if !helpers.Contains(supportedAutogeneratedTypes, lagoonType) {
			autogenEnabled = false
			autogenTLSAcmeEnabled = false
		}

		// check if this service is one that supports backups
		backupsEnabled := false
		if helpers.Contains(typesWithBackups, lagoonType) {
			backupsEnabled = true

		}

		// create the service values
		cService := &ServiceValues{
			Name:                                   composeService,
			OverrideName:                           lagoonOverrideName,
			Type:                                   lagoonType,
			AutogeneratedRoutesEnabled:             autogenEnabled,
			AutogeneratedRoutesTLSAcme:             autogenTLSAcmeEnabled,
			AutogeneratedRoutesRequestVerification: autogeRequestVerification,
			DBaaSEnvironment:                       dbaasEnvironment,
			PersistentVolumePath:                   servicePersistentPath,
			PersistentVolumeName:                   servicePersistentName,
			PersistentVolumeSize:                   servicePersistentSize,
			UseSpotInstances:                       useSpot,
			ForceSpotInstances:                     forceSpot,
			CronjobUseSpotInstances:                cronjobUseSpot,
			CronjobForceSpotInstances:              cronjobForceSpot,
			Replicas:                               spotReplicas,
			InPodCronjobs:                          inpodcronjobs,
			NativeCronjobs:                         nativecronjobs,
			PodSecurityContext:                     buildValues.PodSecurityContext,
			IsDBaaS:                                svcIsDBaaS,
			IsSingle:                               svcIsSingle,
			BackupsEnabled:                         backupsEnabled,
			AdditionalVolumes:                      serviceVolumes,
		}

		// work out the images here and the associated dockerfile and contexts
		// if the type is in the ignored image types, then there is no image to build or pull for this service (eg, its a dbaas service)
		if !helpers.Contains(ignoredImageTypes, lagoonType) {
			imageBuild, err := generateImageBuild(*buildValues, composeServiceValues, composeService)
			if err != nil {
				return nil, err
			}
			cService.ImageBuild = &imageBuild
		}

		// check if the service has a service port override (this only applies to basic(-persistent))
		servicePortOverride := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.service.port")
		if servicePortOverride != "" {
			sPort, err := strconv.Atoi(servicePortOverride)
			if err != nil {
				return nil, fmt.Errorf(
					"the provided service port %s for service %s is not a valid integer: %v",
					servicePortOverride, composeService, err,
				)
			}
			cService.ServicePort = int32(sPort)
		}
		useComposeServices := lagoon.CheckDockerComposeLagoonLabel(composeServiceValues.Labels, "lagoon.service.usecomposeports")
		if useComposeServices == "true" {
			for _, compPort := range composeServiceValues.Ports {
				newService := AdditionalServicePort{
					ServicePort:         compPort,
					ServiceName:         fmt.Sprintf("%s-%d", composeService, compPort.Target),
					ServiceOverrideName: composeService,
				}
				cService.AdditionalServicePorts = append(cService.AdditionalServicePorts, newService)
			}
		}
		return cService, nil
	}
}

package generator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

type Generator struct {
	LagoonYAML          *lagoon.YAML
	BuildValues         *BuildValues
	ActiveEnvironment   *bool
	StandbyEnvironment  *bool
	AutogeneratedRoutes *lagoon.RoutesV2
	MainRoutes          *lagoon.RoutesV2
	ActiveStandbyRoutes *lagoon.RoutesV2
}

type GeneratorInput struct {
	LagoonYAML                 string
	LagoonYAMLOverride         string
	LagoonVersion              string
	BuildName                  string
	SourceRepository           string
	ProjectName                string
	EnvironmentName            string
	EnvironmentType            string
	ActiveEnvironment          string
	StandbyEnvironment         string
	ProjectVariables           string
	EnvironmentVariables       string
	BuildType                  string
	Branch                     string
	GitSHA                     string
	PRNumber                   string
	PRTitle                    string
	PRHeadBranch               string
	PRBaseBranch               string
	PRHeadSHA                  string
	PRBaseSHA                  string
	PromotionSourceEnvironment string
	MonitoringContact          string
	MonitoringStatusPageID     string
	FastlyCacheNoCahce         string
	FastlyAPISecretPrefix      string
	SavedTemplatesPath         string
	ConfigMapSha               string
	BackupConfiguration        BackupConfiguration
	IgnoreNonStringKeyErrors   bool
	IgnoreMissingEnvFiles      bool
	Debug                      bool
	DBaaSClient                *dbaasclient.Client
	ImageReferences            map[string]string
	Namespace                  string
	DefaultBackupSchedule      string
	ImageRegistry              string
	Kubernetes                 string
	CI                         bool
	DynamicSecrets             []string
	DynamicDBaaSSecrets        []string
	ImageCacheBuildArgsJSON    string
	SSHPrivateKey              string
}

func NewGenerator(
	generator GeneratorInput,
) (*Generator, error) {

	// create some initial variables to be passed through the generators
	buildValues := BuildValues{}
	buildValues.FeatureFlags = map[string]bool{}
	lYAML := &lagoon.YAML{}
	autogenRoutes := &lagoon.RoutesV2{}
	mainRoutes := &lagoon.RoutesV2{}
	activeStandbyRoutes := &lagoon.RoutesV2{}

	// environment variables will override what is provided by flags
	// the following variables have been identified as used by custom-ingress objects
	// these are available within a lagoon build as standard
	monitoringContact := helpers.GetEnv("MONITORING_ALERTCONTACT", generator.MonitoringContact, generator.Debug)
	monitoringStatusPageID := helpers.GetEnv("MONITORING_STATUSPAGEID", generator.MonitoringStatusPageID, generator.Debug)
	projectName := helpers.GetEnv("PROJECT", generator.ProjectName, generator.Debug)
	environmentName := helpers.GetEnv("ENVIRONMENT", generator.EnvironmentName, generator.Debug)
	branch := helpers.GetEnv("BRANCH", generator.Branch, generator.Debug)
	prNumber := helpers.GetEnv("PR_NUMBER", generator.PRNumber, generator.Debug)
	prTitle := helpers.GetEnv("PR_TITLE", generator.PRTitle, generator.Debug)
	prHeadBranch := helpers.GetEnv("PR_HEAD_BRANCH", generator.PRHeadBranch, generator.Debug)
	prBaseBranch := helpers.GetEnv("PR_BASE_BRANCH", generator.PRBaseBranch, generator.Debug)
	environmentType := helpers.GetEnv("ENVIRONMENT_TYPE", generator.EnvironmentType, generator.Debug)
	buildType := helpers.GetEnv("BUILD_TYPE", generator.BuildType, generator.Debug)
	activeEnvironment := helpers.GetEnv("ACTIVE_ENVIRONMENT", generator.ActiveEnvironment, generator.Debug)
	standbyEnvironment := helpers.GetEnv("STANDBY_ENVIRONMENT", generator.StandbyEnvironment, generator.Debug)
	fastlyCacheNoCahce := helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", generator.FastlyCacheNoCahce, generator.Debug)
	fastlyAPISecretPrefix := helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", generator.FastlyAPISecretPrefix, generator.Debug)
	lagoonVersion := helpers.GetEnv("LAGOON_VERSION", generator.LagoonVersion, generator.Debug)
	configMapSha := helpers.GetEnv("CONFIG_MAP_SHA", generator.ConfigMapSha, generator.Debug)
	imageRegistry := helpers.GetEnv("REGISTRY", generator.ImageRegistry, generator.Debug)
	kubernetes := helpers.GetEnv("KUBERNETES", generator.Kubernetes, generator.Debug)
	buildName := helpers.GetEnv("LAGOON_BUILD_NAME", generator.BuildName, generator.Debug)
	sourceRepository := helpers.GetEnv("SOURCE_REPOSITORY", generator.SourceRepository, generator.Debug)
	promotionSourceEnvironment := helpers.GetEnv("PROMOTION_SOURCE_ENVIRONMENT", generator.PromotionSourceEnvironment, generator.Debug)
	gitSHA := helpers.GetEnv("LAGOON_GIT_SHA", generator.SourceRepository, generator.Debug)
	prHeadSHA := helpers.GetEnv("PR_HEAD_SHA", generator.PRHeadSHA, generator.Debug)
	prBaseSHA := helpers.GetEnv("PR_BASE_SHA", generator.PRBaseSHA, generator.Debug)
	dynamicSecrets := helpers.GetEnv("DYNAMIC_SECRETS", strings.Join(generator.DynamicSecrets, ","), generator.Debug)
	dynamicDBaaSSecrets := helpers.GetEnv("DYNAMIC_DBAAS_SECRETS", strings.Join(generator.DynamicDBaaSSecrets, ","), generator.Debug)
	imageCacheBuildArgsJSON := helpers.GetEnv("LAGOON_CACHE_BUILD_ARGS", generator.ImageCacheBuildArgsJSON, generator.Debug)
	buildValues.SSHPrivateKey = helpers.GetEnv("SSH_PRIVATE_KEY", generator.SSHPrivateKey, generator.Debug)
	// this is used by CI systems to influence builds, it is rarely used and should probably be abandoned
	buildValues.IsCI = helpers.GetEnvBool("CI", generator.CI, generator.Debug)

	buildValues.ConfigMapSha = configMapSha
	buildValues.BuildName = buildName
	buildValues.Kubernetes = kubernetes
	buildValues.GitSHA = gitSHA
	buildValues.ImageRegistry = imageRegistry
	buildValues.SourceRepository = sourceRepository
	buildValues.PromotionSourceEnvironment = promotionSourceEnvironment
	// get the image references values from the build images output
	buildValues.ImageReferences = generator.ImageReferences
	defaultBackupSchedule := helpers.GetEnv("DEFAULT_BACKUP_SCHEDULE", generator.DefaultBackupSchedule, generator.Debug)
	if defaultBackupSchedule == "" {
		defaultBackupSchedule = "M H(22-2) * * *"
	}

	// try source the namespace from the generator, but whatever is defined in the service account location
	// should be used if one exists, falls back to whatever came in via generator
	namespace := helpers.GetEnv("NAMESPACE", generator.Namespace, generator.Debug)
	namespace, err := helpers.GetNamespace(namespace, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		// a file was found, but there was an issue accessing it
		return nil, err
	}

	buildValues.Backup.K8upVersion = helpers.GetEnv("K8UP_VERSION", generator.BackupConfiguration.K8upVersion, generator.Debug)

	// get the project and environment variables
	projectVariables := helpers.GetEnv("LAGOON_PROJECT_VARIABLES", generator.ProjectVariables, generator.Debug)
	environmentVariables := helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", generator.EnvironmentVariables, generator.Debug)

	// read the .lagoon.yml file and the LAGOON_YAML_OVERRIDE if set
	if err := LoadAndUnmarshalLagoonYml(generator.LagoonYAML, generator.LagoonYAMLOverride, "LAGOON_YAML_OVERRIDE", lYAML, projectName, generator.Debug); err != nil {
		return nil, err
	}
	buildValues.LagoonYAML = *lYAML
	if buildValues.LagoonYAML.EnvironmentVariables.GitSHA == nil || !*buildValues.LagoonYAML.EnvironmentVariables.GitSHA {
		buildValues.GitSHA = "0000000000000000000000000000000000000000"
	}

	//add the dbaas client to build values too
	buildValues.DBaaSClient = generator.DBaaSClient

	buildValues.DefaultBackupSchedule = defaultBackupSchedule

	// set the task scale iterations/wait times
	// these are not user modifiable flags, but are injectable by the controller so individual clusters can
	// set these on their `remote-controller` deployments to be injected to builds.
	buildValues.TaskScaleMaxIterations = helpers.GetEnvInt("LAGOON_FEATURE_FLAG_TASK_SCALE_MAX_ITERATIONS", 30, generator.Debug)
	buildValues.TaskScaleWaitTime = helpers.GetEnvInt("LAGOON_FEATURE_FLAG_TASK_SCALE_WAIT_TIME", 10, generator.Debug)

	// start saving values into the build values variable
	buildValues.Project = projectName
	buildValues.Environment = environmentName
	buildValues.Namespace = namespace
	buildValues.EnvironmentType = environmentType
	buildValues.BuildType = buildType
	buildValues.LagoonVersion = lagoonVersion
	buildValues.ActiveEnvironment = activeEnvironment
	buildValues.StandbyEnvironment = standbyEnvironment
	buildValues.FastlyCacheNoCache = fastlyCacheNoCahce
	buildValues.FastlyAPISecretPrefix = fastlyAPISecretPrefix
	switch buildType {
	case "branch", "promote":
		buildValues.Branch = branch
	case "pullrequest":
		buildValues.PRNumber = prNumber
		buildValues.PRTitle = prTitle
		buildValues.PRHeadBranch = prHeadBranch
		buildValues.PRBaseBranch = prBaseBranch
		buildValues.PRHeadSHA = prHeadSHA
		buildValues.PRBaseSHA = prBaseSHA
		// since pullrequests don't  have a branch
		// we should set the branch to be `pr-PRNUMBER` so that it can be used for matching elsewhere where matching for `branch`
		// using buildvalues is done
		buildValues.Branch = fmt.Sprintf("pr-%v", prNumber)
	}

	// break out of the generator if these requirements are missing
	if projectName == "" || environmentName == "" || environmentType == "" || buildType == "" {
		return nil, fmt.Errorf("missing arguments: project-name, environment-name, environment-type, or build-type not defined")
	}
	switch buildType {
	case "branch", "promote":
		if branch == "" {
			return nil, fmt.Errorf("missing arguments: branch not defined")
		}
	case "pullrequest":
		if prNumber == "" || prHeadBranch == "" || prBaseBranch == "" {
			return nil, fmt.Errorf("missing arguments: pullrequest-number, pullrequest-head-branch, or pullrequest-base-branch not defined")
		}
	}

	// get the dbaas operator http endpoint or fall back to the default
	buildValues.DBaaSOperatorEndpoint = helpers.GetEnv("DBAAS_OPERATOR_HTTP", "http://dbaas.lagoon.svc:5000", generator.Debug)

	// by default, environment routes are not monitored
	buildValues.Monitoring.Enabled = false
	if environmentType == "production" {
		// if this is a production environment, monitoring IS enabled
		buildValues.Monitoring.Enabled = true
		buildValues.Monitoring.AlertContact = monitoringContact
		buildValues.Monitoring.StatusPageID = monitoringStatusPageID
		// check if the environment is active or standby
		if environmentName == activeEnvironment {
			buildValues.IsActiveEnvironment = true
		}
		if environmentName == standbyEnvironment {
			buildValues.IsStandbyEnvironment = true
		}
	}

	// handle the dynamic secret volume creation from input secret names
	if dynamicSecrets != "" {
		for _, ds := range strings.Split(dynamicSecrets, ",") {
			buildValues.DynamicSecretMounts = append(buildValues.DynamicSecretMounts, DynamicSecretMounts{
				Name:      fmt.Sprintf("dynamic-%s", ds),
				MountPath: fmt.Sprintf("/var/run/secrets/lagoon/dynamic/%s", ds),
				ReadOnly:  true,
			})
			buildValues.DynamicSecretVolumes = append(buildValues.DynamicSecretVolumes, DynamicSecretVolumes{
				Name: fmt.Sprintf("dynamic-%s", ds),
				Secret: DynamicSecret{
					SecretName: ds,
					Optional:   false,
				},
			})
		}
	}
	if dynamicDBaaSSecrets != "" {
		// if there are any dynamic dbaas secrets defined, send them here
		buildValues.DynamicDBaaSSecrets = strings.Split(dynamicDBaaSSecrets, ",")
	}

	// unmarshal and then merge the two so there is only 1 set of variables to iterate over
	projectVars := []lagoon.EnvironmentVariable{}
	envVars := []lagoon.EnvironmentVariable{}
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	mergedVariables := lagoon.MergeVariables(projectVars, envVars)
	// collect a bunch of the default LAGOON_X based build variables that are injected into `lagoon-env` and make them available
	configVars := collectBuildVariables(buildValues)
	// add the calculated build runtime variables into the existing variable slice
	// this will later be used to add `runtime|global` scope into the `lagoon-env` configmap
	buildValues.EnvironmentVariables = lagoon.MergeVariables(mergedVariables, configVars)

	// if the core version is provided from the API, set the buildvalues LagoonVersion to this instead
	lagoonCoreVersion, _ := lagoon.GetLagoonVariable("LAGOON_SYSTEM_CORE_VERSION", []string{"internal_system"}, buildValues.EnvironmentVariables)
	if lagoonCoreVersion != nil {
		buildValues.LagoonVersion = lagoonCoreVersion.Value
	}

	// handle generating the container registry login generation here, extract from the `.lagoon.yml` firstly
	if err := configureContainerRegistries(&buildValues); err != nil {
		return nil, err
	}

	// check for readwritemany to readwriteonce flag, disabled by default
	rwx2rwo := CheckFeatureFlag("RWX_TO_RWO", buildValues.EnvironmentVariables, generator.Debug)
	if rwx2rwo == "enabled" {
		buildValues.RWX2RWO = true
	}

	// check for isolation network policy, disabled by default
	isolationNetworkPolicy := CheckFeatureFlag("ISOLATION_NETWORK_POLICY", buildValues.EnvironmentVariables, generator.Debug)
	if isolationNetworkPolicy == "enabled" {
		buildValues.IsolationNetworkPolicy = true
	}

	// check for imagecache override, disabled by default
	imageCache := CheckFeatureFlag("IMAGECACHE_REGISTRY", buildValues.EnvironmentVariables, generator.Debug)
	if imageCache != "" {
		// strip the scheme, only provide the host
		u, _ := url.Parse(imageCache)
		if u.Host == "" {
			imageCache = fmt.Sprintf("%s/", imageCache)
		} else {
			imageCache = fmt.Sprintf("%s/", u.Host)
		}
		buildValues.ImageCache = imageCache
	}

	// check route quota
	lagoonRouteQuota, _ := lagoon.GetLagoonVariable("LAGOON_ROUTE_QUOTA", []string{"internal_system"}, buildValues.EnvironmentVariables)
	if lagoonRouteQuota != nil {
		routeQuota, err := strconv.Atoi(lagoonRouteQuota.Value)
		if err != nil {
			return nil, fmt.Errorf("route quota does not convert to integer, contact your Lagoon administrator")
		}
		buildValues.RouteQuota = &routeQuota
	}

	// check the environment for INGRESS_CLASS flag, will be "" if there are none found
	ingressClass := CheckFeatureFlag("INGRESS_CLASS", buildValues.EnvironmentVariables, generator.Debug)
	buildValues.IngressClass = ingressClass

	// check for rootless workloads
	rootlessWorkloads := CheckFeatureFlag("ROOTLESS_WORKLOAD", buildValues.EnvironmentVariables, generator.Debug)
	if rootlessWorkloads == "enabled" {
		buildValues.FeatureFlags["rootlessworkloads"] = true
		buildValues.PodSecurityContext = PodSecurityContext{
			RunAsGroup: 0,
			RunAsUser:  10000,
			FsGroup:    10001,
		}
	}

	fsOnRootMismatch := CheckFeatureFlag("FS_ON_ROOT_MISMATCH", buildValues.EnvironmentVariables, generator.Debug)
	if fsOnRootMismatch == "enabled" {
		buildValues.PodSecurityContext.OnRootMismatch = true
	}

	// check admin features for resources
	buildValues.Resources.Limits.Memory = CheckAdminFeatureFlag("CONTAINER_MEMORY_LIMIT", false)
	buildValues.Resources.Limits.EphemeralStorage = CheckAdminFeatureFlag("EPHEMERAL_STORAGE_LIMIT", false)
	buildValues.Resources.Requests.EphemeralStorage = CheckAdminFeatureFlag("EPHEMERAL_STORAGE_REQUESTS", false)
	// validate that what is provided
	if buildValues.Resources.Limits.Memory != "" {
		err := ValidateResourceQuantity(buildValues.Resources.Limits.Memory)
		if err != nil {
			return nil, fmt.Errorf("provided memory limit %s is not a valid resource quantity", buildValues.Resources.Limits.Memory)
		}
	}
	if buildValues.Resources.Limits.EphemeralStorage != "" {
		err := ValidateResourceQuantity(buildValues.Resources.Limits.EphemeralStorage)
		if err != nil {
			return nil, fmt.Errorf("provided ephemeral storage limit %s is not a valid resource quantity", buildValues.Resources.Limits.EphemeralStorage)
		}
	}
	if buildValues.Resources.Requests.EphemeralStorage != "" {
		err := ValidateResourceQuantity(buildValues.Resources.Requests.EphemeralStorage)
		if err != nil {
			return nil, fmt.Errorf("provided  ephemeral storage requests %s is not a valid resource quantity", buildValues.Resources.Requests.EphemeralStorage)
		}
	}

	// get any variables from the API here that could be used to influence a build or services within the environment
	// collect docker buildkit value
	dockerBuildKit, _ := lagoon.GetLagoonVariable("DOCKER_BUILDKIT", []string{"build"}, buildValues.EnvironmentVariables)
	if dockerBuildKit != nil {
		buildValues.DockerBuildKit, _ = strconv.ParseBool(dockerBuildKit.Value)
	}

	// get any lagoon service type overrides
	lagoonServiceTypes, _ := lagoon.GetLagoonVariable("LAGOON_SERVICE_TYPES", nil, buildValues.EnvironmentVariables)
	buildValues.ServiceTypeOverrides = lagoonServiceTypes

	// get any dbaas environment type overrides
	lagoonDBaaSEnvironmentTypes, _ := lagoon.GetLagoonVariable("LAGOON_DBAAS_ENVIRONMENT_TYPES", nil, buildValues.EnvironmentVariables)
	buildValues.DBaaSEnvironmentTypeOverrides = lagoonDBaaSEnvironmentTypes

	// check autogenerated routes for fastly `LAGOON_FEATURE_FLAG(_FORCE|_DEFAULT)_FASTLY_AUTOGENERATED` using feature flags
	// @TODO: eventually deprecate fastly functionality in favour of a more generic implementation
	autogeneratedRoutesFastly := CheckFeatureFlag("FASTLY_AUTOGENERATED", buildValues.EnvironmentVariables, generator.Debug)
	if autogeneratedRoutesFastly == "enabled" {
		buildValues.AutogeneratedRoutesFastly = true
	} else {
		buildValues.AutogeneratedRoutesFastly = false
	}
	// check legacy variable in envvars
	// @TODO: eventually deprecate fastly functionality in favour of a more generic implementation
	lagoonAutogeneratedFastly, _ := lagoon.GetLagoonVariable("LAGOON_FASTLY_AUTOGENERATED", nil, buildValues.EnvironmentVariables)
	if lagoonAutogeneratedFastly != nil {
		if lagoonAutogeneratedFastly.Value == "enabled" {
			buildValues.AutogeneratedRoutesFastly = true
		} else {
			buildValues.AutogeneratedRoutesFastly = false
		}
	}
	// check legacy variable in envvars
	cronjobsDisabled, _ := lagoon.GetLagoonVariable("LAGOON_CRONJOBS_DISABLED", nil, buildValues.EnvironmentVariables)
	if cronjobsDisabled != nil {
		if cronjobsDisabled.Value == "true" {
			buildValues.CronjobsDisabled = true
		} else {
			buildValues.CronjobsDisabled = false
		}
	}

	// @TODO: eventually fail builds if this is not set https://github.com/uselagoon/build-deploy-tool/issues/56
	// lagoonDBaaSFallbackSingle, _ := lagoon.GetLagoonVariable("LAGOON_FEATURE_FLAG_DBAAS_FALLBACK_SINGLE", nil, buildValues.EnvironmentVariables)
	// buildValues.DBaaSFallbackSingle = helpers.StrToBool(lagoonDBaaSFallbackSingle.Value)

	/* start backups configuration */
	err = generateBackupValues(&buildValues, buildValues.EnvironmentVariables, generator.Debug)
	if err != nil {
		return nil, err
	}
	/* end backups configuration */

	/*
		start compose->service configuration
		!! IMPORTANT !!
		build values should be calculated as much as possible before being passed to the generate services function
	*/
	err = generateServicesFromDockerCompose(&buildValues, generator.IgnoreNonStringKeyErrors, generator.IgnoreMissingEnvFiles, generator.Debug)
	if err != nil {
		return nil, err
	}

	if imageCacheBuildArgsJSON != "" {
		err = json.Unmarshal([]byte(imageCacheBuildArgsJSON), &buildValues.ImageCacheBuildArguments)
		if err != nil {
			return nil, err
		}
	}
	buildValues.ImageBuildArguments = collectImageBuildArguments(buildValues)
	/* end compose->service configuration */

	/* start route generation */
	// create all the routes for this environment and store the primary and secondary routes into values
	// populate the autogenRoutes, mainRoutes and activeStandbyRoutes here and load them
	buildValues.Route, buildValues.Routes, buildValues.AutogeneratedRoutes, err = generateRoutes(
		buildValues.EnvironmentVariables,
		buildValues,
		autogenRoutes,
		mainRoutes,
		activeStandbyRoutes,
		generator.Debug,
	)
	if err != nil {
		return nil, err
	}
	if buildValues.RouteQuota != nil {
		customRoutes := len(buildValues.Routes) - len(buildValues.AutogeneratedRoutes)
		if customRoutes > *buildValues.RouteQuota && *buildValues.RouteQuota != -1 {
			return nil, fmt.Errorf("this environment requests %d custom routes, this would exceed the route quota of %d", customRoutes, *buildValues.RouteQuota)
		}
	}
	/* end route generation configuration */

	// finally return the generator values, this should be a mostly complete version of the resulting data needed for a build
	// another step will collect the current or known state of a build.
	// the output of the generator and the output of that state collector will eventually replace a lot of the legacy BASH script
	return &Generator{
		BuildValues:         &buildValues,
		ActiveEnvironment:   &buildValues.IsActiveEnvironment,
		StandbyEnvironment:  &buildValues.IsStandbyEnvironment,
		AutogeneratedRoutes: autogenRoutes,
		MainRoutes:          mainRoutes,
		ActiveStandbyRoutes: activeStandbyRoutes,
	}, nil
}

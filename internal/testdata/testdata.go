package testdata

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/machinery/utils/namespace"
)

// basic data structure for test data using the generator
type TestData struct {
	AlertContact               string
	StatusPageID               string
	BuildName                  string
	SourceRepository           string
	Kubernetes                 string
	ProjectName                string
	EnvironmentName            string
	Branch                     string
	GitSHA                     string
	PRNumber                   string
	PRTitle                    string
	PRHeadBranch               string
	PRBaseBranch               string
	PRHeadSHA                  string
	PRBaseSHA                  string
	EnvironmentType            string
	BuildType                  string
	ActiveEnvironment          string
	StandbyEnvironment         string
	CacheNoCache               string
	ServiceID                  string
	SecretPrefix               string
	IngressClass               string
	ProjectVars                string
	EnvVars                    string
	ProjectVariables           []lagoon.EnvironmentVariable
	EnvVariables               []lagoon.EnvironmentVariable
	LagoonVersion              string
	LagoonYAML                 string
	ValuesFilePath             string
	K8UPVersion                string
	DefaultBackupSchedule      string
	ControllerDevSchedule      string
	ControllerPRSchedule       string
	Namespace                  string
	ImageReferences            map[string]string
	ConfigMapSha               string
	ImageRegistry              string
	PromotionSourceEnvironment string
	PrivateRegistryURLS        []string
	DynamicSecrets             []string
	DynamicDBaaSSecrets        []string
	ImageCacheBuildArgsJSON    string
	SSHPrivateKey              string
	BuildPodVariables          []helpers.EnvironmentVariable
}

// helper function to set up all the environment variables from provided testdata
func SetupEnvironment(rootCmd cobra.Command, templatePath string, t TestData) (generator.GeneratorInput, error) {
	err := os.Setenv("MONITORING_ALERTCONTACT", t.AlertContact)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("MONITORING_STATUSPAGEID", t.StatusPageID)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PROJECT", t.ProjectName)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("ENVIRONMENT", t.EnvironmentName)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("BRANCH", t.Branch)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_NUMBER", t.PRNumber)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_TITLE", t.PRTitle)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_HEAD_BRANCH", t.PRHeadBranch)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_BASE_BRANCH", t.PRBaseBranch)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("ENVIRONMENT_TYPE", t.EnvironmentType)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("BUILD_TYPE", t.BuildType)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("ACTIVE_ENVIRONMENT", t.ActiveEnvironment)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("STANDBY_ENVIRONMENT", t.StandbyEnvironment)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", t.CacheNoCache)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	pv, _ := json.Marshal(t.ProjectVariables)
	err = os.Setenv("LAGOON_PROJECT_VARIABLES", string(pv))
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	ev, _ := json.Marshal(t.EnvVariables)
	err = os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", string(ev))
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_VERSION", t.LagoonVersion)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_FEATURE_FLAG_DEFAULT_INGRESS_CLASS", t.IngressClass)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", t.ControllerDevSchedule)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_FEATURE_BACKUP_PR_SCHEDULE", t.ControllerPRSchedule)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("DEFAULT_BACKUP_SCHEDULE", t.DefaultBackupSchedule)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("REGISTRY", t.ImageRegistry)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("SOURCE_REPOSITORY", t.SourceRepository)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_BUILD_NAME", t.BuildName)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("KUBERNETES", t.Kubernetes)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_HEAD_SHA", t.PRHeadSHA)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PR_BASE_SHA", t.PRBaseSHA)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_GIT_SHA", t.GitSHA)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("PROMOTION_SOURCE_ENVIRONMENT", t.PromotionSourceEnvironment)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("DYNAMIC_SECRETS", strings.Join(t.DynamicSecrets, ","))
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("DYNAMIC_DBAAS_SECRETS", strings.Join(t.DynamicDBaaSSecrets, ","))
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("LAGOON_CACHE_BUILD_ARGS", t.ImageCacheBuildArgsJSON)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	err = os.Setenv("SSH_PRIVATE_KEY", t.SSHPrivateKey)
	if err != nil {
		return generator.GeneratorInput{}, err
	}
	// this can b used to pass OS level variables likw what would be present in a build pod when it
	// is started by the remote-controller
	for _, osv := range t.BuildPodVariables {
		err = os.Setenv(osv.Name, osv.Value)
		if err != nil {
			return generator.GeneratorInput{}, err
		}
	}

	generator, err := generator.GenerateInput(rootCmd, false)
	if err != nil {
		return generator, err
	}
	generator.LagoonYAML = t.LagoonYAML
	generator.ImageReferences = t.ImageReferences
	generator.ConfigMapSha = t.ConfigMapSha
	generator.SavedTemplatesPath = templatePath
	// add dbaasclient overrides for tests
	generator.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
		RetryMax:     5,
		RetryWaitMin: time.Duration(10) * time.Millisecond,
		RetryWaitMax: time.Duration(50) * time.Millisecond,
	})

	generator.Namespace = namespace.GenerateNamespaceName("", t.EnvironmentName, t.ProjectName, "", "lagoon", false)

	generator.BackupConfiguration.K8upVersion = t.K8UPVersion

	return generator, nil
}

func GetSeedData(t TestData, defaultProjectVariables bool) TestData {
	// set up the default values, but all values are overwriteable via the input
	rt := TestData{
		AlertContact:    "alertcontact", // will be deprecated eventually
		StatusPageID:    "statuspageid", // will be deprecated eventually
		ProjectName:     "example-project",
		EnvironmentType: "production",
		BuildType:       "branch",
		LagoonVersion:   "v2.7.x",
		ProjectVariables: []lagoon.EnvironmentVariable{
			{
				Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
				Value: "${service}-${project}-${environment}.example.com",
				Scope: "internal_system",
			},
		},
		K8UPVersion:      "v1",
		ConfigMapSha:     "abcdefg1234567890",
		ImageRegistry:    "harbor.example",
		BuildName:        "lagoon-build-abcdefg",
		SourceRepository: "ssh://git@example.com/lagoon-demo.git",
		Kubernetes:       "remote-cluster1",
		GitSHA:           "abcdefg123456",
		SSHPrivateKey:    "-----BEGIN OPENSSH PRIVATE KEY-----\nthisisafakekey\n-----END OPENSSH PRIVATE KEY-----",
	}
	if t.ProjectName != "" {
		rt.ProjectName = t.ProjectName
	}
	if t.EnvironmentName != "" {
		rt.EnvironmentName = t.EnvironmentName
	}
	if t.Branch != "" {
		rt.Branch = t.Branch
	}
	if t.EnvironmentType != "" {
		rt.EnvironmentType = t.EnvironmentType
	}
	if t.BuildType != "" {
		rt.BuildType = t.BuildType
	}
	if rt.BuildType == "promote" {
		rt.PromotionSourceEnvironment = "promote-main"
	}
	if t.PRNumber != "" {
		rt.PRNumber = t.PRNumber
	}
	if t.PRTitle != "" {
		rt.PRTitle = t.PRTitle
	}
	if t.PRHeadBranch != "" {
		rt.PRHeadBranch = t.PRHeadBranch
	}
	if t.PRBaseBranch != "" {
		rt.PRBaseBranch = t.PRBaseBranch
	}
	if t.PRHeadSHA != "" {
		rt.PRHeadSHA = t.PRHeadSHA
	}
	if t.PRBaseSHA != "" {
		rt.PRBaseSHA = t.PRBaseSHA
	}
	if t.LagoonVersion != "" {
		rt.LagoonVersion = t.LagoonVersion
	}
	if t.LagoonYAML != "" {
		rt.LagoonYAML = t.LagoonYAML
	}
	if t.ProjectVariables != nil && defaultProjectVariables {
		rt.ProjectVariables = append(rt.ProjectVariables, t.ProjectVariables...)
	} else if !defaultProjectVariables {
		rt.ProjectVariables = t.ProjectVariables
	}
	if t.EnvVariables != nil {
		rt.EnvVariables = append(rt.EnvVariables, t.EnvVariables...)
	}
	if t.ActiveEnvironment != "" {
		rt.ActiveEnvironment = t.ActiveEnvironment
	}
	if t.StandbyEnvironment != "" {
		rt.StandbyEnvironment = t.StandbyEnvironment
	}
	if t.IngressClass != "" {
		rt.IngressClass = t.IngressClass
	}
	if t.K8UPVersion != "" {
		rt.K8UPVersion = t.K8UPVersion
	}
	if t.DefaultBackupSchedule != "" {
		rt.DefaultBackupSchedule = t.DefaultBackupSchedule
	}
	if t.ControllerDevSchedule != "" {
		rt.ControllerDevSchedule = t.ControllerDevSchedule
	}
	if t.ControllerPRSchedule != "" {
		rt.ControllerPRSchedule = t.ControllerPRSchedule
	}
	if t.ImageReferences != nil {
		rt.ImageReferences = t.ImageReferences
	}
	if t.ConfigMapSha != "" {
		rt.ConfigMapSha = t.ConfigMapSha
	}
	// will be deprecated eventually
	if t.AlertContact != "" {
		rt.AlertContact = t.AlertContact
	}
	if t.StatusPageID != "" {
		rt.StatusPageID = t.StatusPageID
	}
	if t.PrivateRegistryURLS != nil {
		rt.PrivateRegistryURLS = t.PrivateRegistryURLS
	}
	if t.DynamicSecrets != nil {
		rt.DynamicSecrets = t.DynamicSecrets
	}
	if t.DynamicDBaaSSecrets != nil {
		rt.DynamicDBaaSSecrets = t.DynamicDBaaSSecrets
	}
	if t.ImageCacheBuildArgsJSON != "" {
		rt.ImageCacheBuildArgsJSON = t.ImageCacheBuildArgsJSON
	}
	if t.SSHPrivateKey != "" {
		rt.SSHPrivateKey = t.SSHPrivateKey
	}
	if t.BuildPodVariables != nil {
		rt.BuildPodVariables = t.BuildPodVariables
	}
	return rt
}

package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var templateGitCredential = &cobra.Command{
	Use:     "git-credential",
	Aliases: []string{"gitcred", "gc"},
	Short:   "Create a git credential file",
	RunE: func(cmd *cobra.Command, args []string) error {
		credFile, err := cmd.Flags().GetString("credential-file")
		if err != nil {
			return fmt.Errorf("error reading credential-file flag: %v", err)
		}
		geninput, err := GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		gitURL, err := TemplateGitCredential(geninput, credFile)
		if err != nil {
			return err
		}
		fmt.Println(gitURL)
		return nil
	},
}

// TemplateGitCredential creates a git credential secret if required
func TemplateGitCredential(geninput generator.GeneratorInput, file string) (string, error) {
	// get the source repository from the build pod environment variables
	sourceRepository := helpers.GetEnv("SOURCE_REPOSITORY", "", false)
	projectName := helpers.GetEnv("PROJECT", "", false)

	// because this runs before anything has been checked out in the repo, we have to query variables directly
	// and handle any merging that is usually done in generator
	variables := generator.GetLagoonEnvVars()
	lagoonYAML := lagoon.YAML{}
	if err := generator.LoadAndUnmarshalLagoonYml(geninput.LagoonYAML, geninput.LagoonYAMLOverride, "LAGOON_YAML_OVERRIDE", &lagoonYAML, projectName, false); err != nil {
		return "", err
	}
	// parse the repository into a url struct so the username and password can be injected into http/https based urls
	u, err := url.Parse(sourceRepository)
	if err != nil {
		return "", fmt.Errorf("unable to parse provided gitUrl")
	}
	// if this is a http or https based url, it may need a username and password
	if helpers.Contains([]string{"http", "https"}, u.Scheme) {
		// only create the git credential if the source repo is a http/https repostiory
		urls, err := configureGitCredentials(lagoonYAML, variables)
		if err != nil {
			return "", err
		}
		var s string
		for idx, url := range urls {
			if idx == 0 {
				s = url.String()
			} else {
				s = fmt.Sprintf("%s\n%s", s, url.String())
			}
		}
		err = os.WriteFile(file, []byte(s), 0644)
		if err != nil {
			return "", err
		}
		if urls == nil {
			return "", nil
		}
		// return the new url only if it was modified with a username and password
		return "store", nil
	}
	// otherwise return whatever was provided to the build
	return "", nil
}

func init() {
	templateCmd.AddCommand(templateGitCredential)
	templateGitCredential.Flags().String("credential-file", "", "The file to store the credential in")
}

// this converts lagoon.yml git credentials definitions into a git credential file
// that is used to perform git pulls on private repositories using https
func configureGitCredentials(lagoonYAML lagoon.YAML, variables []lagoon.EnvironmentVariable) ([]url.URL, error) {
	urls := []url.URL{}
	for n, cr := range lagoonYAML.GitCredentials {
		// check for an repository password
		password, _ := lagoon.GetBuildVariable(fmt.Sprintf("GITREPO_%s_PASSWORD", n), variables)
		// if not found
		if password == nil {
			return nil, fmt.Errorf("no username defined for git repository %s, expected 'build' scope variable named %s", n, fmt.Sprintf("GITREPO_%s_PASSWORD", n))
		}
		// check for an repository username
		username, _ := lagoon.GetBuildVariable(fmt.Sprintf("GITREPO_%s_USERNAME", n), variables)
		// if not found
		if username == nil {
			return nil, fmt.Errorf("no username defined for git repository %s, expected 'build' scope variable named %s", n, fmt.Sprintf("GITREPO_%s_USERNAME", n))
		}
		if cr.URL == "" {
			// if no url defined, assume github
			cr.URL = "https://github.com"
		}
		u, _ := url.Parse(cr.URL)
		// craft the url for the git credential
		u.User = url.UserPassword(username.Value, password.Value)
		urls = append(urls, url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
			User:   u.User,
		})
	}
	return urls, nil
}

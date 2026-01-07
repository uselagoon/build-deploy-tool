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
		gitURL, err := TemplateGitCredential(credFile)
		if err != nil {
			return err
		}
		fmt.Println(gitURL)
		return nil
	},
}

// TemplateGitCredential creates a git credential secret if required
func TemplateGitCredential(file string) (string, error) {
	// get the source repository from the build pod environment variables
	sourceRepository := helpers.GetEnv("SOURCE_REPOSITORY", "", false)

	// because this runs before anything has been checked out in the repo, we have to query variables directly
	// and handle any merging that is usually done in generator
	variables := generator.GetLagoonEnvVars()

	// parse the repository into a url struct so the username and password can be injected into http/https based urls
	u, err := url.Parse(sourceRepository)
	if err != nil {
		return "", fmt.Errorf("unable to parse provided gitUrl")
	}
	cred := url.URL{}
	// if this is a http or https based url, it may need a username and password
	if helpers.Contains([]string{"http", "https"}, u.Scheme) {
		// since the provided source repository could be public or private, check for the `LAGOON_GIT_HTTPS_X`
		// variables, ignore errors for these lookups as they're in most cases going to be "not found"
		username, _ := lagoon.GetLagoonVariable("LAGOON_GIT_HTTPS_USERNAME", []string{"build"}, variables)
		password, _ := lagoon.GetLagoonVariable("LAGOON_GIT_HTTPS_PASSWORD", []string{"build"}, variables)
		if username != nil && password == nil {
			return "", fmt.Errorf("LAGOON_GIT_HTTPS_USERNAME was provided, but not LAGOON_GIT_HTTPS_PASSWORD")
		}
		if username == nil && password != nil {
			return "", fmt.Errorf("LAGOON_GIT_HTTPS_PASSWORD was provided, but not LAGOON_GIT_HTTPS_USERNAME")
		}
		// if both are found, set the user auth into the url
		if username != nil && password != nil {
			u.User = url.UserPassword(username.Value, password.Value)
			cred.Scheme = u.Scheme
			cred.Host = u.Host
			cred.User = u.User
			err := os.WriteFile(file, []byte(cred.String()), 0644)
			if err != nil {
				return "", err
			}
			// return the new url only if it was modified with a username and password
			return "store", nil
		}
	}
	// otherwise return whatever was provided to the build
	return "", nil
}

func init() {
	templateCmd.AddCommand(templateGitCredential)
	templateGitCredential.Flags().String("credential-file", "", "The file to store the credential in")
}

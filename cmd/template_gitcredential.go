package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

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

	// because this runs before anything has been checked out in the repo, we have to query variables directly
	// and handle any merging that is usually done in generator
	variables := generator.GetLagoonEnvVars()
	// parse the repository into a url struct so the username and password can be injected into http/https based urls
	u, err := url.Parse(sourceRepository)
	if err != nil {
		return "", fmt.Errorf("unable to parse provided gitUrl")
	}
	// if this is a http or https based url, it may need a username and password
	if helpers.Contains([]string{"http", "https"}, u.Scheme) {
		// only create the git credential if the source repo is a http/https repostiory
		urls, err := configureGitCredentials(variables)
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
func configureGitCredentials(variables []lagoon.EnvironmentVariable) ([]url.URL, error) {
	urls := []url.URL{}
	hasGithub := false
	for _, cr := range variables {
		if strings.Contains(cr.Name, "GITREPO_GITHUB_USERNAME") || strings.Contains(cr.Name, "GITREPO_GITHUB_PASSWORD") && !hasGithub {
			url, err := extractValues("GITHUB", "https://github.com", variables)
			if err != nil {
				return nil, err
			}
			urls = append(urls, url)
			hasGithub = true
			continue
		}
		// check for any GITREPO_x_URL variables to determine if non github repo is defined
		if strings.HasPrefix(cr.Name, "GITREPO_") && strings.HasSuffix(cr.Name, "_URL") {
			repoName := strings.TrimPrefix(cr.Name, "GITREPO_")
			repoName = strings.TrimSuffix(repoName, "_URL")
			rURL, _ := lagoon.GetBuildVariable(fmt.Sprintf("GITREPO_%s_URL", repoName), variables)
			repoURL := rURL.Value
			url, err := extractValues(repoName, repoURL, variables)
			if err != nil {
				return nil, err
			}
			urls = append(urls, url)
			continue
		}
	}
	return urls, nil
}

func extractValues(repoName, repoURL string, variables []lagoon.EnvironmentVariable) (url.URL, error) {
	// check for a repository password
	password, _ := lagoon.GetBuildVariable(fmt.Sprintf("GITREPO_%s_PASSWORD", repoName), variables)
	// if not found
	if password == nil {
		return url.URL{}, fmt.Errorf("no username defined for git repository %s, expected 'build' scope variable named %s", repoName, fmt.Sprintf("GITREPO_%s_PASSWORD", repoName))
	}
	// check for a repository username
	username, _ := lagoon.GetBuildVariable(fmt.Sprintf("GITREPO_%s_USERNAME", repoName), variables)
	// if not found
	if username == nil {
		return url.URL{}, fmt.Errorf("no username defined for git repository %s, expected 'build' scope variable named %s", repoName, fmt.Sprintf("GITREPO_%s_USERNAME", repoName))
	}
	u, _ := url.Parse(repoURL)
	// craft the url for the git credential
	u.User = url.UserPassword(username.Value, password.Value)
	return url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		User:   u.User,
	}, nil
}

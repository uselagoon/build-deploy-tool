package generator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	machinerynamespace "github.com/uselagoon/machinery/utils/namespace"
	"k8s.io/apimachinery/pkg/util/validation"
)

// this converts lagoon.yml container registry definitions into build values container registry definitions
// that are then used to generate secrets, or get passed to docker login commands within the build
func configureContainerRegistries(buildValues *BuildValues) error {
	for n, cr := range buildValues.LagoonYAML.ContainerRegistries {
		// check for required inputs
		if cr.Username == "" {
			return fmt.Errorf("no username defined for registry %s", n)
		}
		if cr.Password == "" {
			return fmt.Errorf("no password defined for registry %s", n)
		}
		// check for an override password
		password, _ := lagoon.GetLagoonVariable(fmt.Sprintf("REGISTRY_%s_PASSWORD", n), []string{"container_registry"}, buildValues.EnvironmentVariables)
		passwordSource := fmt.Sprintf("Lagoon API environment variable %s", fmt.Sprintf("REGISTRY_%s_PASSWORD", n))
		if password == nil {
			// no override found, check for a variable that matches the name in the password field
			password, _ = lagoon.GetLagoonVariable(cr.Password, []string{"container_registry"}, buildValues.EnvironmentVariables)
			passwordSource = fmt.Sprintf("Lagoon API environment variable %s", cr.Password)
		}
		if password == nil {
			// finally, if no password is found in any variables, pass in the one from the `.lagoon.yml` directly
			password = &lagoon.EnvironmentVariable{Value: cr.Password}
			passwordSource = ".lagoon.yml (we recommend using an environment variable, see the docs on container-registries for more information)"
		}
		// check for an override username
		username, _ := lagoon.GetLagoonVariable(fmt.Sprintf("REGISTRY_%s_USERNAME", n), []string{"container_registry"}, buildValues.EnvironmentVariables)
		usernameSource := fmt.Sprintf("Lagoon API environment variable %s", fmt.Sprintf("REGISTRY_%s_USERNAME", n))
		if username == nil {
			username = &lagoon.EnvironmentVariable{Value: cr.Username}
			usernameSource = ".lagoon.yml"
		}
		eru := cr.URL
		u, _ := url.Parse(eru)
		if u.Host == "" {
			eru = fmt.Sprintf("%s", eru)
		} else {
			eru = fmt.Sprintf("%s", u.Host)
		}
		// truncate the secret name to fit within the DNS1123subdomain spec before creating it
		secretName := fmt.Sprintf("lagoon-private-registry-%s", machinerynamespace.MakeSafe(n))
		if err := validation.IsDNS1123Subdomain(strings.ToLower(secretName)); err != nil {
			secretName = fmt.Sprintf("%s-%s", secretName[:len(secretName)-10], helpers.GetMD5HashWithNewLine(machinerynamespace.MakeSafe(n))[:5])
		}
		buildValues.ContainerRegistry = append(buildValues.ContainerRegistry, ContainerRegistry{
			Name:           n,
			Username:       username.Value,
			Password:       password.Value,
			URL:            eru,
			UsernameSource: usernameSource,
			PasswordSource: passwordSource,
			SecretName:     secretName,
		})
	}
	return nil
}

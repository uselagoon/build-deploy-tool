package lagoon

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	authSnippet          = "nginx.ingress.kubernetes.io/auth-snippet"
	configurationSnippet = "nginx.ingress.kubernetes.io/configuration-snippet"
	modsecuritySnippet   = "nginx.ingress.kubernetes.io/modsecurity-snippet"
	serverSnippet        = "nginx.ingress.kubernetes.io/server-snippet"
	streamSnippet        = "nginx.ingress.kubernetes.io/stream-snippet"
	useRegex             = "nginx.ingress.kubernetes.io/use-regex"
	traefikAnnotations   = "traefik.ingress.kubernetes.io"
)

// validSnippets is the allow-list of snippets that Lagoon will accept.
// Currently these are only valid in server-snippet and configuration-snippet
// annotations.
var validSnippets = regexp.MustCompile(
	`^(rewrite +[^; ]+ +[^; ]+( (last|break|redirect|permanent))?;|` +
		`add_header +([^; ]+|"[^"]+"|'[^']+') +([^; ]+|"[^"]+"|'[^']+')( always)?;|` +
		`set_real_ip_from +[^; ]+;|` +
		`more_set_headers +(-s +("[^"]+"|'[^']+')|-t +("[^"]+"|'[^']+')|("[^"]+"|'[^']+'))+;|` +
		` )+$`)

// validate returns true if the annotations are valid, and false otherwise.
func validate(annotations map[string]string, r *regexp.Regexp,
	annotation string) (string, bool) {
	if ss, ok := annotations[annotation]; ok {
		for _, line := range strings.Split(ss, "\n") {
			line = strings.TrimSpace(line)
			if len(line) > 0 && !r.MatchString(line) {
				return line, false
			}
		}
	}
	return "", true
}

// validateRouteAnnotations returns an error if the annotations on the routes in an environment
// are invalid, and nil otherwise.
func validateRouteAnnotations(yamlRoutes RoutesV2) error {
	for _, lagoonRoute := range yamlRoutes.Routes {
		// auth-snippet
		if _, ok := lagoonRoute.Annotations[authSnippet]; ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				authSnippet, lagoonRoute.Domain,
				"this annotation is restricted")
		}
		// configuration-snippet
		if annotation, ok := validate(lagoonRoute.Annotations, validSnippets,
			configurationSnippet); !ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				configurationSnippet, lagoonRoute.Domain, annotation)
		}
		// modsecurity-snippet
		if _, ok := lagoonRoute.Annotations[modsecuritySnippet]; ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				modsecuritySnippet, lagoonRoute.Domain,
				"this annotation is restricted")
		}
		// server-snippet
		if annotation, ok := validate(lagoonRoute.Annotations, validSnippets,
			serverSnippet); !ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				serverSnippet, lagoonRoute.Domain, annotation)
		}
		// stream-snippet
		if _, ok := lagoonRoute.Annotations[streamSnippet]; ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				streamSnippet, lagoonRoute.Domain, "this annotation is restricted")
		}
		// use-regex
		if _, ok := lagoonRoute.Annotations[useRegex]; ok {
			return fmt.Errorf(
				"invalid %s annotation on route %s: %s",
				useRegex, lagoonRoute.Domain, "this annotation is restricted")
		}
		// restrict all traefik annotations
		for annotation := range lagoonRoute.Annotations {
			if strings.Contains(annotation, traefikAnnotations) {
				return fmt.Errorf(
					"invalid %s annotation on route %s: %s",
					annotation, lagoonRoute.Domain,
					"this annotation is restricted")

			}
		}
	}
	return nil
}

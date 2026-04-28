package generator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type CORSMiddleware struct {
	AllowOrigins     string `json:"allowOrigins,omitempty"`
	AllowMethods     string `json:"allowMethods,omitempty"`
	AllowHeaders     string `json:"allowHeaders,omitempty"`
	AllowCredentials *bool  `json:"allowCredentials,omitempty"`
	ExposeHeaders    string `json:"exposeHeaders,omitempty"`
	MaxAge           *int   `json:"maxAge,omitempty"`
}

type RealIPFromMiddleware struct {
	ExcludedNets []string `json:"excludednets"`
}

func generateMiddleware(buildValues *BuildValues, mainRoutes *lagoon.RoutesV2) error {
	buildValues.TraefikMiddlewares = make(map[string]traefik.MiddlewareSpec)
	// add the aergia idling middleware for traefik
	buildValues.TraefikMiddlewares["aergia"] = traefik.MiddlewareSpec{
		Errors: &traefik.ErrorPage{
			Status: []string{"503"},
			Query:  fmt.Sprintf("/?namespace=%s&url={url}", buildValues.Namespace),
			Service: traefik.Service{
				LoadBalancerSpec: traefik.LoadBalancerSpec{
					Name:      "aergia-backend",
					Namespace: "aergia",
					Port: intstr.IntOrString{
						IntVal: 80,
					},
				},
			},
		},
	}
	buildValues.TraefikMiddlewares["https-redirect"] = traefik.MiddlewareSpec{
		RedirectScheme: &dynamic.RedirectScheme{
			Scheme: "https",
		},
	}
	buildValues.TraefikMiddlewares["x-robots"] = traefik.MiddlewareSpec{
		Headers: &dynamic.Headers{
			CustomResponseHeaders: map[string]string{
				"X-Robots-Tag": "noindex, nofollow",
			},
		},
	}
	// can't do pod names, so just put a more generic traefik x-lagoon to the namespace
	buildValues.TraefikMiddlewares["x-lagoon"] = traefik.MiddlewareSpec{
		Headers: &dynamic.Headers{
			CustomResponseHeaders: map[string]string{
				"X-Lagoon": fmt.Sprintf("%s>traefik>%s", buildValues.Kubernetes, buildValues.Namespace),
			},
		},
	}
	for idx, route := range mainRoutes.Routes {
		// placeholder for headers for this route
		// build hsts middleware
		routeHeaders := &dynamic.Headers{}
		containsHeaders := false
		if route.HSTSEnabled != nil && *route.HSTSEnabled {
			if route.HSTSIncludeSubdomains != nil {
				routeHeaders.STSIncludeSubdomains = *route.HSTSIncludeSubdomains
				containsHeaders = true
			}
			if route.HSTSMaxAge != 0 {
				sts := int64(route.HSTSMaxAge)
				routeHeaders.STSSeconds = &sts
				containsHeaders = true
			}
			if route.HSTSPreload != nil {
				routeHeaders.STSPreload = *route.HSTSPreload
				containsHeaders = true
			}
		}
		// build ipallowlist middleware
		whitelistRange, hasWhitelist := route.Annotations["nginx.ingress.kubernetes.io/whitelist-source-range"]
		denylistRange, hasDenylist := route.Annotations["nginx.ingress.kubernetes.io/denylist-source-range"]
		if hasWhitelist && hasDenylist {
			return fmt.Errorf("cannot set whitelist-source-range and denylist-source-range, conflict on route %s", route.Domain)
		}
		if hasWhitelist {
			mainRoutes.Routes[idx].HasIPAllowList = true
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-ipallowlist", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				IPWhiteList: &dynamic.IPWhiteList{
					SourceRange: splitTrim(whitelistRange),
				},
			}
		}
		if hasDenylist {
			mainRoutes.Routes[idx].HasIPAllowList = true
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-ipallowlist", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				IPWhiteList: &dynamic.IPWhiteList{
					SourceRange: []string{"0.0.0.0/0"},
					IPStrategy: &dynamic.IPStrategy{
						ExcludedIPs: splitTrim(denylistRange),
					},
				},
			}
		}
		// handle permanent/temporary redirect
		permRedirect, hasPermRedirect := route.Annotations["nginx.ingress.kubernetes.io/permanent-redirect"]
		tempRedirect, hasTempRedirect := route.Annotations["nginx.ingress.kubernetes.io/temporal-redirect"]
		if hasPermRedirect && hasTempRedirect {
			return fmt.Errorf("cannot set permanent-redirect and temporal-redirect, conflict on route %s", route.Domain)
		}
		if hasPermRedirect || hasTempRedirect {
			mainRoutes.Routes[idx].HasRedirect = true
			redirect := traefik.MiddlewareSpec{
				RedirectRegex: &dynamic.RedirectRegex{
					Regex:     fmt.Sprintf("^https?://%s/(.*)", route.Domain),
					Permanent: false,
				},
			}
			if hasPermRedirect {
				redirect.RedirectRegex.Replacement = strings.ReplaceAll(permRedirect, "$request_uri", "/${1}")
				redirect.RedirectRegex.Permanent = true
			}
			if hasTempRedirect {
				redirect.RedirectRegex.Replacement = strings.ReplaceAll(tempRedirect, "$request_uri", "/${1}")
			}
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-redirect", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = redirect
		}
		// scan server snippet for cases
		if value, ok := route.Annotations["nginx.ingress.kubernetes.io/server-snippet"]; ok {
			// handle set_real_ip_from in server-snippet
			ips := serverSnippetSetRealIPFrom(value)
			if ips != nil {
				mainRoutes.Routes[idx].HasSetRealIPFrom = true
				realIPFrom := RealIPFromMiddleware{
					ExcludedNets: ips,
				}
				realIPFromBytes, _ := json.Marshal(realIPFrom)
				buildValues.TraefikMiddlewares[fmt.Sprintf("%s-setrealip", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
					Plugin: map[string]v1.JSON{
						"traefik-real-ip": {
							Raw: realIPFromBytes,
						},
					},
				}
			}

			headers := serverSnippetAddHeader(value)
			if len(headers) > 0 {
				// add any headers from the server-snippet to the custom response headers
				routeHeaders.CustomResponseHeaders = headers
				containsHeaders = true
			}
		}
		// handle cors settings
		if value, ok := route.Annotations["nginx.ingress.kubernetes.io/enable-cors"]; ok && value == "true" {
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]; ok {
				routeHeaders.AccessControlAllowOriginList = splitTrim(value)
				containsHeaders = true
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-methods"]; ok {
				routeHeaders.AccessControlAllowMethods = splitTrim(value)
				containsHeaders = true
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-headers"]; ok {
				routeHeaders.AccessControlAllowHeaders = splitTrim(value)
				containsHeaders = true
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"]; ok {
				valBool, _ := strconv.ParseBool(value)
				routeHeaders.AccessControlAllowCredentials = valBool
				containsHeaders = true
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-expose-headers"]; ok {
				routeHeaders.AccessControlExposeHeaders = splitTrim(value)
				containsHeaders = true
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-max-age"]; ok {
				valInt, _ := strconv.Atoi(value)
				routeHeaders.AccessControlMaxAge = int64(valInt)
				containsHeaders = true
			}
		}
		// handle basic auth
		if value, ok := route.Annotations["nginx.ingress.kubernetes.io/auth-type"]; ok && value == "basic" {
			mainRoutes.Routes[idx].HasBasicAuth = true
			basicAuth := traefik.MiddlewareSpec{
				BasicAuth: &traefik.BasicAuth{},
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/auth-secret"]; ok {
				basicAuth.BasicAuth.Secret = value
				// secret must be a `kubernetes.io/basic-auth` type with `username` and `password` keys
				// not the same as nginx with `auth` key only
				// lagoon could support more generic basic auth protection with this method than
				// the current method only supported in nginx images, means more people could use it
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/auth-realm"]; ok {
				basicAuth.BasicAuth.Realm = value
			}
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-basicauth", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = basicAuth
		}
		if containsHeaders {
			mainRoutes.Routes[idx].HasHeaders = true
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-headers", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				Headers: routeHeaders,
			}
		}
	}
	return nil
}

var realIPRegex = regexp.MustCompile(`(?i)set_real_ip_from\s+([^;\s#]+)`)

// extract set_real_ip_from values from nginx-ingress server snippet and turn them into a slice of ip addresses
func serverSnippetSetRealIPFrom(snippet string) []string {
	var ips []string
	seen := map[string]struct{}{}
	scanner := bufio.NewScanner(strings.NewReader(snippet))
	for scanner.Scan() {
		line := scanner.Text()
		// ignore comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		match := realIPRegex.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}
		ip := strings.TrimSpace(match[1])
		if _, exists := seen[ip]; !exists {
			seen[ip] = struct{}{}
			ips = append(ips, ip)
		}
	}
	return ips
}

var addHeaderRegex = regexp.MustCompile(`add_header\s+([A-Za-z0-9-]+)\s+(?:"([^"]*)"|([^;]+))\s*(?:always)?\s*;`)

// extract add_header from server-snippets to add to headers
func serverSnippetAddHeader(conf string) map[string]string {
	headers := map[string]string{}
	matches := addHeaderRegex.FindAllStringSubmatch(conf, -1)
	for _, m := range matches {
		name := m[1]
		value := m[2]
		if value == "" {
			value = strings.TrimSpace(m[3])
		}
		headers[name] = value
	}
	return headers
}

// split and trip a string
func splitTrim(input string) []string {
	slc := strings.Split(input, ",")
	for i := range slc {
		slc[i] = strings.TrimSpace(slc[i])
	}
	return slc
}

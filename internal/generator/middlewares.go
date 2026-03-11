package generator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
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
		// build hsts middleware
		if route.HSTSEnabled != nil && *route.HSTSEnabled {
			stsHeader := &dynamic.Headers{}
			if route.HSTSIncludeSubdomains != nil {
				stsHeader.STSIncludeSubdomains = *route.HSTSIncludeSubdomains
			}
			if route.HSTSMaxAge != 0 {
				stsHeader.STSSeconds = int64(route.HSTSMaxAge)
			}
			if route.HSTSPreload != nil {
				stsHeader.STSPreload = *route.HSTSPreload
			}
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-hsts", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				Headers: stsHeader,
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
					SourceRange: strings.Split(whitelistRange, ","),
				},
			}
		}
		if hasDenylist {
			mainRoutes.Routes[idx].HasIPAllowList = true
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-ipallowlist", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				IPWhiteList: &dynamic.IPWhiteList{
					SourceRange: []string{"0.0.0.0/0"},
					IPStrategy: &dynamic.IPStrategy{
						ExcludedIPs: strings.Split(denylistRange, ","),
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
		}
		// handle cors settings
		if value, ok := route.Annotations["nginx.ingress.kubernetes.io/enable-cors"]; ok && value == "true" {
			mainRoutes.Routes[idx].HasCORS = true
			corsMiddleware := CORSMiddleware{}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]; ok {
				corsMiddleware.AllowOrigins = value
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-methods"]; ok {
				corsMiddleware.AllowMethods = value
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-headers"]; ok {
				corsMiddleware.AllowHeaders = value
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"]; ok {
				valBool, _ := strconv.ParseBool(value)
				corsMiddleware.AllowCredentials = &valBool
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-expose-headers"]; ok {
				corsMiddleware.ExposeHeaders = value
			}
			if value, ok := route.Annotations["nginx.ingress.kubernetes.io/cors-max-age"]; ok {
				valInt, _ := strconv.Atoi(value)
				corsMiddleware.MaxAge = &valInt
			}
			corsByte, _ := json.Marshal(corsMiddleware)
			buildValues.TraefikMiddlewares[fmt.Sprintf("%s-cors", helpers.GetBase32EncodedLowercase(helpers.GetSha256Hash(route.IngressName))[:8])] = traefik.MiddlewareSpec{
				Plugin: map[string]v1.JSON{
					"corsmiddleware": {
						Raw: corsByte,
					},
				},
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

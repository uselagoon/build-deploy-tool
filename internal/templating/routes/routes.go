package routes

import (
	"github.com/uselagoon/lagoon-routegen/internal/helpers"
	"github.com/uselagoon/lagoon-routegen/internal/lagoon"
)

// GenerateRouteStructure generate the route structure for lagoon routes.
func GenerateRouteStructure(genRoutes *lagoon.RoutesV2, routeMap map[string][]lagoon.Route, variables []lagoon.EnvironmentVariable, activeStandby bool) {
	for rName, lagoonRoutes := range routeMap {
		for _, lagoonRoute := range lagoonRoutes {
			newRoute := &lagoon.RouteV2{}
			newRoute.TLSAcme = helpers.BoolPtr(true)
			newRoute.Insecure = helpers.StrPtr("Redirect")
			newRoute.MonitoringPath = "/"
			newRoute.HSTS = helpers.StrPtr("null")
			newRoute.Annotations = map[string]string{}
			newRoute.Fastly.ServiceID = ""
			newRoute.Fastly.Watch = false
			if activeStandby {
				newRoute.Migrate = helpers.BoolPtr(true)
			}
			if lagoonRoute.Name == "" {
				// this route has fields
				for iName, ingress := range lagoonRoute.Ingresses {
					newRoute.Domain = iName
					newRoute.Service = rName
					newRoute.Fastly = ingress.Fastly
					if ingress.Annotations != nil {
						newRoute.Annotations = ingress.Annotations
					}
					if ingress.TLSAcme != nil {
						newRoute.TLSAcme = ingress.TLSAcme
					}
					if ingress.Insecure != nil {
						newRoute.Insecure = ingress.Insecure
					}
					if ingress.HSTS != nil {
						newRoute.HSTS = ingress.HSTS
					}
				}
			} else {
				// this route is just a domain
				newRoute.Domain = lagoonRoute.Name
				newRoute.Service = rName
			}
			fConfig, err := lagoon.GenerateFastlyConfiguration("", newRoute.Fastly.ServiceID, newRoute.Domain, variables)
			if err != nil {
			}
			newRoute.Fastly = fConfig

			genRoutes.Routes = append(genRoutes.Routes, *newRoute)
		}
	}
}

// MergeRouteStructures merge route structures for lagoon routes.
func MergeRouteStructures(genRoutes lagoon.RoutesV2, apiRoutes lagoon.RoutesV2) lagoon.RoutesV2 {
	finalRoutes := &lagoon.RoutesV2{}
	existsInAPI := false
	// replace any routes from the lagoon yaml with ones from the api
	// this only modifies ones that exist in lagoon yaml
	for _, route := range genRoutes.Routes {
		add := lagoon.RouteV2{}
		for _, aRoute := range apiRoutes.Routes {
			if aRoute.Domain == route.Domain {
				existsInAPI = true
				add = aRoute
				add.Fastly = aRoute.Fastly
				if aRoute.TLSAcme != nil {
					add.TLSAcme = aRoute.TLSAcme
				} else {
					add.TLSAcme = helpers.BoolPtr(true)
				}
				if aRoute.Insecure != nil {
					add.Insecure = aRoute.Insecure
				} else {
					add.Insecure = helpers.StrPtr("Redirect")
				}
				if aRoute.HSTS != nil {
					add.HSTS = aRoute.HSTS
				} else {
					add.HSTS = helpers.StrPtr("null")
				}
				if aRoute.Annotations != nil {
					add.Annotations = aRoute.Annotations
				} else {
					add.Annotations = map[string]string{}
				}
			}
		}
		if existsInAPI {
			finalRoutes.Routes = append(finalRoutes.Routes, add)
			existsInAPI = false
		} else {
			finalRoutes.Routes = append(finalRoutes.Routes, route)
		}
	}
	// add any that exist in the api only to the final routes list
	for _, aRoute := range apiRoutes.Routes {
		add := lagoon.RouteV2{}
		for _, route := range finalRoutes.Routes {
			add = aRoute
			add.Fastly = aRoute.Fastly
			if aRoute.TLSAcme != nil {
				add.TLSAcme = aRoute.TLSAcme
			} else {
				add.TLSAcme = helpers.BoolPtr(true)
			}
			if aRoute.Insecure != nil {
				add.Insecure = aRoute.Insecure
			} else {
				add.Insecure = helpers.StrPtr("Redirect")
			}
			if aRoute.HSTS != nil {
				add.HSTS = aRoute.HSTS
			} else {
				add.HSTS = helpers.StrPtr("null")
			}
			if aRoute.Annotations != nil {
				add.Annotations = aRoute.Annotations
			} else {
				add.Annotations = map[string]string{}
			}
			if aRoute.Domain == route.Domain {
				existsInAPI = true
			}
		}
		if existsInAPI {
			existsInAPI = false
		} else {
			finalRoutes.Routes = append(finalRoutes.Routes, add)
		}
	}
	return *finalRoutes
}

package templating

import (
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

// LinkedServiceCalculator checks the provided services to see if there are any linked services
// linked services are mostly just `nginx-php` but lagoon has the possibility to support more than this in the future
func LinkedServiceCalculator(services []generator.ServiceValues) []generator.ServiceValues {
	linkedMap := make(map[string][]generator.ServiceValues)
	retServices := []generator.ServiceValues{}
	linkedOrder := []string{}

	// go over the services twice to extract just the linked services (the override names will be the same in a linked service)
	for _, s1 := range services {
		for _, s2 := range services {
			if s1.OverrideName == s2.OverrideName && s1.Name != s2.Name {
				linkedMap[s1.OverrideName] = append(linkedMap[s1.OverrideName], s1)
				linkedOrder = helpers.AppendIfMissing(linkedOrder, s1.OverrideName)
			}
		}
	}
	// go over the services again and any that are in the services that aren't in the linked map (again the override name is the key)
	// add it as a standalone service
	for _, s1 := range services {
		if _, ok := linkedMap[s1.OverrideName]; !ok {
			retServices = append(retServices, s1)
		}
	}

	// go over the linked services and add the linkedservice to the main service
	// example would be adding the `php` service in docker-compose to the `nginx` service as a `LinkedService` definition
	// this allows the generated service values to carry across
	for _, name := range linkedOrder {
		service := generator.ServiceValues{}
		if len(linkedMap[name]) == 2 {
			for idx, s := range linkedMap[name] {
				if idx == 0 {
					service = s
				}
				if idx == 1 {
					service.LinkedService = &s
				}
			}
		}
		// then add it to the slice of services to return
		retServices = append(retServices, service)
	}
	return retServices
}

package lagoon

import (
	"encoding/json"
	"reflect"
	"strconv"
)

// Ingress represents a Lagoon route.
type Ingress struct {
	TLSAcme        *bool             `json:"tls-acme,omitempty"`
	Migrate        *bool             `json:"migrate,omitempty"`
	Insecure       *string           `json:"insecure,omitempty"`
	HSTS           *string           `json:"hsts,omitempty"`
	MonitoringPath string            `json:"monitoring-path,omitempty"`
	Fastly         Fastly            `json:"fastly,omitempty"`
	Annotations    map[string]string `json:"annotations,omitempty"`
}

// Annotations .
type Annotations struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Fastly represents the fastly configuration for a Lagoon route
type Fastly struct {
	ServiceID     string `json:"service-id,omitempty"`
	APISecretName string `json:"api-secret-name,omitempty"`
	Watch         bool   `json:"watch,omitempty"`
}

// Route can be either a string or a map[string]Ingress, so we must
// implement a custom unmarshaller.
type Route struct {
	Name      string
	Ingresses map[string]Ingress
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Route) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Name); err == nil {
		return nil
	}
	if err := json.Unmarshal(data, &r.Ingresses); err != nil {
		// some things in .lagoon.yml can be defined as a bool or string and lagoon builds don't care
		// but types are more strict, so this unmarshaler attempts to change between the two types
		// that can be bool or string
		tmpMap := map[string]interface{}{}
		json.Unmarshal(data, &tmpMap)
		for k := range tmpMap {
			if _, ok := tmpMap[k].(map[string]interface{})["tls-acme"]; ok {
				if reflect.TypeOf(tmpMap[k].(map[string]interface{})["tls-acme"]).Kind() == reflect.String {
					vBool, err := strconv.ParseBool(tmpMap[k].(map[string]interface{})["tls-acme"].(string))
					if err == nil {
						tmpMap[k].(map[string]interface{})["tls-acme"] = vBool
					}
				}
			}
			if _, ok := tmpMap[k].(map[string]interface{})["fastly"]; ok {
				if reflect.TypeOf(tmpMap[k].(map[string]interface{})["fastly"].(map[string]interface{})["watch"]).Kind() == reflect.String {
					vBool, err := strconv.ParseBool(tmpMap[k].(map[string]interface{})["fastly"].(map[string]interface{})["watch"].(string))
					if err == nil {
						tmpMap[k].(map[string]interface{})["fastly"].(map[string]interface{})["watch"] = vBool
					}
				}
			}
		}
		newData, _ := json.Marshal(tmpMap)
		return json.Unmarshal(newData, &r.Ingresses)
	}
	return json.Unmarshal(data, &r.Ingresses)
}

// Environment represents a Lagoon environment.
type Environment struct {
	Routes []map[string][]Route `json:"routes"`
}

// ProductionRoutes represents an active/standby configuration.
type ProductionRoutes struct {
	Active  *Environment `json:"active"`
	Standby *Environment `json:"standby"`
}

// YAML represents the .lagoon.yml file.
type YAML struct {
	Environments     map[string]Environment `json:"environments"`
	ProductionRoutes *ProductionRoutes      `json:"production_routes"`
}

// Environments .
type Environments map[string]Environment

// EnvironmentVariable is used to define Lagoon environment variables.
type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

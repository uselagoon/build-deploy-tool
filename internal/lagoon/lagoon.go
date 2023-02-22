package lagoon

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"

	"github.com/imdario/mergo"
	"sigs.k8s.io/yaml"
)

// ProductionRoutes represents an active/standby configuration.
type ProductionRoutes struct {
	Active  *Environment `json:"active"`
	Standby *Environment `json:"standby"`
}

// Cronjob represents a Lagoon cronjob.
type Cronjob struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// Environment represents a Lagoon environment.
type Environment struct {
	AutogenerateRoutes *bool                `json:"autogenerateRoutes"`
	Types              map[string]string    `json:"types"`
	Routes             []map[string][]Route `json:"routes"`
	Cronjobs           []Cronjob            `json:"cronjobs"`
}

// Environments .
type Environments map[string]Environment

// TaskRun .
type TaskRun struct {
	Run Task `json:"run"`
}

// Tasks .
type Tasks struct {
	Prerollout  []TaskRun `json:"pre-rollout"`
	Postrollout []TaskRun `json:"post-rollout"`
}

// YAML represents the .lagoon.yml file.
type YAML struct {
	DockerComposeYAML string            `json:"docker-compose-yaml"`
	Environments      Environments      `json:"environments"`
	ProductionRoutes  *ProductionRoutes `json:"production_routes"`
	Tasks             Tasks             `json:"tasks"`
	Routes            Routes            `json:"routes"`
	BackupRetention   BackupRetention   `json:"backup-retention"`
	BackupSchedule    BackupSchedule    `json:"backup-schedule"`
}

type BackupRetention struct {
	Production Retention `json:"production"`
}

type BackupSchedule struct {
	Production string `json:"production"`
}

type Retention struct {
	Hourly  *int `json:"hourly"`
	Daily   *int `json:"daily"`
	Weekly  *int `json:"weekly"`
	Monthly *int `json:"monthly"`
}

// Routes .
type Routes struct {
	Autogenerate Autogenerate `json:"autogenerate"`
}

// Autogenerate .
type Autogenerate struct {
	Enabled           *bool    `json:"enabled"`
	AllowPullRequests *bool    `json:"allowPullRequests"`
	Insecure          string   `json:"insecure"`
	Prefixes          []string `json:"prefixes"`
	TLSAcme           *bool    `json:"tls-acme,omitempty"`
	IngressClass      string   `json:"ingressClass"`
}

func (a *Routes) UnmarshalJSON(data []byte) error {
	tmpMap := map[string]interface{}{}
	json.Unmarshal(data, &tmpMap)
	if value, ok := tmpMap["autogenerate"]; ok {
		// @TODO: eventually lagoon should be more strict, but in lagoonyaml version 2 we could do this
		// some things in .lagoon.yml can be defined as a bool or string and lagoon builds don't care
		// but types are more strict, so this unmarshaler attempts to change between the two types
		// that can be bool or string
		if _, ok := value.(map[string]interface{})["tls-acme"]; ok {
			if reflect.TypeOf(value.(map[string]interface{})["tls-acme"]).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(value.(map[string]interface{})["tls-acme"].(string))
				if err == nil {
					value.(map[string]interface{})["tls-acme"] = vBool
				}
			}
		}
		if _, ok := value.(map[string]interface{})["enabled"]; ok {
			if reflect.TypeOf(value.(map[string]interface{})["enabled"]).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(value.(map[string]interface{})["enabled"].(string))
				if err == nil {
					value.(map[string]interface{})["enabled"] = vBool
				}
			}
		}
		if _, ok := value.(map[string]interface{})["allowPullRequests"]; ok {
			if reflect.TypeOf(value.(map[string]interface{})["allowPullRequests"]).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(value.(map[string]interface{})["allowPullRequests"].(string))
				if err == nil {
					value.(map[string]interface{})["allowPullRequests"] = vBool
				}
			}
		}
		newData, _ := json.Marshal(value)
		return json.Unmarshal(newData, &a.Autogenerate)
	}
	return nil
}

// UnmarshalLagoonYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshalLagoonYAML(file string, l *YAML, p *map[string]interface{}) error {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", file, err)
	}
	// lagoon.yml
	err = yaml.Unmarshal(rawYAML, l)
	if err != nil {
		return err
	}
	// polysite
	err = yaml.Unmarshal(rawYAML, p)
	if err != nil {
		return err
	}
	return nil
}

func MergeLagoonYAMLs(destination *YAML, source *YAML) error {
	if err := mergeLagoonYAMLTasks(&destination.Tasks.Prerollout, &source.Tasks.Prerollout); err != nil {
		return err
	}
	if err := mergeLagoonYAMLTasks(&destination.Tasks.Postrollout, &source.Tasks.Postrollout); err != nil {
		return err
	}
	sortLagoonYamlTasksByWeight(destination.Tasks.Prerollout)
	sortLagoonYamlTasksByWeight(destination.Tasks.Postrollout)
	return nil
}

func sortLagoonYamlTasksByWeight(tasks []TaskRun) {
	sort.Slice(tasks, func(i int, j int) bool {
		return tasks[i].Run.Weight < tasks[j].Run.Weight
	})
}

func mergeLagoonYAMLTasks(left *[]TaskRun, right *[]TaskRun) error {
	for i, rightTask := range *right {
		appendToLeft := true
		for j, leftTask := range *left {
			if leftTask.Run.Name != "" && leftTask.Run.Name == rightTask.Run.Name {
				//here we merge the two, rather than appending
				appendToLeft = false
				if err := mergo.Merge(&(*left)[j].Run, &(*right)[i].Run, mergo.WithOverride); err != nil {
					return err
				}

			}
		}
		if appendToLeft {
			*left = append(*left, rightTask)
		}
	}
	return nil
}

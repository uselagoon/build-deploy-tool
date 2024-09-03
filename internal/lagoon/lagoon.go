package lagoon

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"

	"dario.cat/mergo"
	"sigs.k8s.io/yaml"
)

// ProductionRoutes represents an active/standby configuration.
type ProductionRoutes struct {
	Active  *Environment `json:"active"`
	Standby *Environment `json:"standby"`
}

// Environment represents a Lagoon environment.
type Environment struct {
	AutogenerateRoutes *bool                `json:"autogenerateRoutes"`
	Types              map[string]string    `json:"types"`
	Routes             []map[string][]Route `json:"routes"`
	Cronjobs           []Cronjob            `json:"cronjobs"`
	Overrides          map[string]Override  `json:"overrides,omitempty"`
}

// Cronjob represents a Lagoon cronjob.
type Cronjob struct {
	Name     string `json:"name"`
	Service  string `json:"service"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	InPod    *bool  `json:"inPod"`
}

type Override struct {
	Build Build  `json:"build,omitempty"`
	Image string `json:"image,omitempty"`
}

type Build struct {
	Dockerfile string `json:"dockerfile,omitempty"`
	Context    string `json:"context,omitempty"`
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
	DockerComposeYAML    string                       `json:"docker-compose-yaml"`
	Environments         Environments                 `json:"environments"`
	ProductionRoutes     *ProductionRoutes            `json:"production_routes"`
	Tasks                Tasks                        `json:"tasks"`
	Routes               Routes                       `json:"routes"`
	BackupRetention      BackupRetention              `json:"backup-retention"`
	BackupSchedule       BackupSchedule               `json:"backup-schedule"`
	EnvironmentVariables EnvironmentVariables         `json:"environment_variables,omitempty"`
	ContainerRegistries  map[string]ContainerRegistry `json:"container-registries,omitempty"`
}

type ContainerRegistry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

type EnvironmentVariables struct {
	GitSHA *bool `json:"git_sha"`
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
	Enabled             *bool    `json:"enabled"`
	AllowPullRequests   *bool    `json:"allowPullRequests"`
	Insecure            string   `json:"insecure"`
	Prefixes            []string `json:"prefixes"`
	TLSAcme             *bool    `json:"tls-acme,omitempty"`
	IngressClass        string   `json:"ingressClass"`
	RequestVerification *bool    `json:"disableRequestVerification,omitempty"`
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
					// @TODO: add warning functionality here to inform that users should fix their yaml to be boolean not string
					// this could warn in a yaml validation step at the start of builds
					value.(map[string]interface{})["tls-acme"] = vBool
				}
			}
		}
		if _, ok := value.(map[string]interface{})["enabled"]; ok {
			if reflect.TypeOf(value.(map[string]interface{})["enabled"]).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(value.(map[string]interface{})["enabled"].(string))
				if err == nil {
					// @TODO: add warning functionality here to inform that users should fix their yaml to be boolean not string
					// this could warn in a yaml validation step at the start of builds
					value.(map[string]interface{})["enabled"] = vBool
				}
			}
		}
		if _, ok := value.(map[string]interface{})["allowPullRequests"]; ok {
			if reflect.TypeOf(value.(map[string]interface{})["allowPullRequests"]).Kind() == reflect.String {
				vBool, err := strconv.ParseBool(value.(map[string]interface{})["allowPullRequests"].(string))
				if err == nil {
					// @TODO: add warning functionality here to inform that users should fix their yaml to be boolean not string
					// this could warn in a yaml validation step at the start of builds
					value.(map[string]interface{})["allowPullRequests"] = vBool
				}
			}
		}
		newData, _ := json.Marshal(value)
		return json.Unmarshal(newData, &a.Autogenerate)
	}
	return nil
}

func (a *EnvironmentVariables) UnmarshalJSON(data []byte) error {
	tmpMap := map[string]interface{}{}
	json.Unmarshal(data, &tmpMap)
	if value, ok := tmpMap["git_sha"]; ok {
		// @TODO: eventually lagoon should be more strict, but in lagoonyaml version 2 we could do this
		// some things in .lagoon.yml can be defined as a bool or string and lagoon builds don't care
		// but types are more strict, so this unmarshaler attempts to change between the two types
		// that can be bool or string
		if reflect.TypeOf(value).Kind() == reflect.String {
			vBool, err := strconv.ParseBool(value.(string))
			if err == nil {
				// @TODO: add warning functionality here to inform that users should fix their yaml to be boolean not string
				// this could warn in a yaml validation step at the start of builds
				value = vBool
			}
		}
		newData, _ := json.Marshal(value)
		return json.Unmarshal(newData, &a.GitSHA)
	}
	return nil
}

// UnmarshalLagoonYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshalLagoonYAML(file string, l *YAML, project string) error {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", file, err)
	}
	// lagoon.yml
	err = yaml.Unmarshal(rawYAML, l)
	if err != nil {
		return err
	}

	// if this is a polysite, then unmarshal the polysite data into a normal lagoon environments yaml
	// this is done so that all other generators only need to know how to interact with one type of environment
	p := map[string]interface{}{}
	err = yaml.Unmarshal(rawYAML, &p)
	if err != nil {
		return err
	}
	// get all the cronjobs from the top level environments cronjobs into a new map
	polycrons := map[string][]Cronjob{}
	for en, e := range l.Environments {
		polycrons[en] = e.Cronjobs
	}
	if _, ok := p[project]; ok {
		s, err := yaml.Marshal(p[project])
		if err != nil {
			return err
		}
		// this step copies the polysite environments block over the yaml effectively wiping it clean of anything else it may have had
		err = yaml.Unmarshal(s, l)
		if err != nil {
			return err
		}
	}
	// iterate over the new environments (from the polysite lagoonyml $project.Environments)
	for en, e := range l.Environments {
		// check if polysite crons exist
		val := polycrons[en]
		// check if the two aren't already the same, no need to do anything otherwise
		if !reflect.DeepEqual(e.Cronjobs, val) {
			if len(e.Cronjobs) == 0 {
				// if there are no cronjobs from the polysite project, set the cronjobs to be the older top level cronjobs only
				e.Cronjobs = val
			} else {
				for _, c1 := range e.Cronjobs {
					for _, c2 := range val {
						// check if the original top level cronjobs exist
						if c1.Name == c2.Name {
							continue
						}
						e.Cronjobs = append(e.Cronjobs, c2)
					}
				}
			}
			l.Environments[en] = e
		}
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

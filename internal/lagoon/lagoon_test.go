package lagoon

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"sigs.k8s.io/yaml"
)

func TestLagoonYAMLUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want *YAML
	}{
		{
			name: "test-booleans-represented-as-strings",
			yaml: "test-resources/lagoon-stringbooleans.yml",
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				Environments: Environments{
					"master": Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"a.example.com": {
												TLSAcme: helpers.BoolPtr(true),
											},
										},
									},
									{
										Name: "b.example.com",
									},
									{
										Name: "c.example.com",
									},
								},
							},
						},
					},
				},
				ProductionRoutes: &ProductionRoutes{
					Active: &Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"active.example.com": {
												TLSAcme:  helpers.BoolPtr(true),
												Insecure: helpers.StrPtr("Redirect"),
											},
										},
									},
								},
							},
						},
					},
					Standby: &Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"standby.example.com": {
												TLSAcme:  helpers.BoolPtr(false),
												Insecure: helpers.StrPtr("Redirect"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test-booleans-represented-as-booleans",
			yaml: "test-resources/lagoon-booleans.yml",
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				Environments: Environments{
					"master": Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"a.example.com": {
												TLSAcme: helpers.BoolPtr(true),
											},
										},
									},
									{
										Name: "b.example.com",
									},
									{
										Name: "c.example.com",
									},
								},
							},
						},
					},
				},
				ProductionRoutes: &ProductionRoutes{
					Active: &Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"active.example.com": {
												TLSAcme:  helpers.BoolPtr(true),
												Insecure: helpers.StrPtr("Redirect"),
											},
										},
									},
								},
							},
						},
					},
					Standby: &Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Ingresses: map[string]Ingress{
											"standby.example.com": {
												TLSAcme:  helpers.BoolPtr(false),
												Insecure: helpers.StrPtr("Redirect"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawYAML, err := os.ReadFile(tt.yaml)
			if err != nil {
				panic(fmt.Errorf("couldn't read %v: %v", tt.yaml, err))
			}
			got := &YAML{}
			yaml.Unmarshal(rawYAML, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = got %v, want %v", got, tt.want)
			}
		})
	}
}

package lagoon

import (
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

func TestUnmarshalLagoonYAML(t *testing.T) {
	type args struct {
		file string
		l    *YAML
		p    *map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *YAML
		wantErr bool
	}{
		{
			name: "test-booleans-represented-as-strings",
			args: args{
				file: "../../test-resources/lagoon-yaml/test1/lagoon.yml",
				l:    &YAML{},
				p:    &map[string]interface{}{},
			},
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				Environments: Environments{
					"main": Environment{
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
			args: args{
				file: "../../test-resources/lagoon-yaml/test2/lagoon.yml",
				l:    &YAML{},
				p:    &map[string]interface{}{},
			},
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				Environments: Environments{
					"main": Environment{
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
			name: "test-booleans-represented-as-strings-and-booleans",
			args: args{
				file: "../../test-resources/lagoon-yaml/test3/lagoon.yml",
				l:    &YAML{},
				p:    &map[string]interface{}{},
			},
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				Environments: Environments{
					"main": Environment{
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
			name: "test-backup-retention",
			args: args{
				file: "../../test-resources/lagoon-yaml/test4/lagoon.yml",
				l:    &YAML{},
				p:    &map[string]interface{}{},
			},
			want: &YAML{
				DockerComposeYAML: "docker-compose.yml",
				BackupRetention: BackupRetention{
					Production: Retention{
						Hourly:  helpers.IntPtr(0),
						Daily:   helpers.IntPtr(10),
						Weekly:  helpers.IntPtr(6),
						Monthly: helpers.IntPtr(2),
					},
				},
				BackupSchedule: BackupSchedule{
					Production: "M/15 5 * * 0",
				},
				Environments: Environments{
					"main": Environment{
						Routes: []map[string][]Route{
							{
								"nginx": {
									{
										Name: "a.example.com",
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
			if err := UnmarshalLagoonYAML(tt.args.file, tt.args.l, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalLagoonYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.l, tt.want) {
				t.Errorf("Unmarshal() = got %v, want %v", tt.args.l, tt.want)
			}
		})
	}
}

func TestMergeLagoonYAMLs(t *testing.T) {
	type args struct {
		left  *YAML
		right *YAML
	}
	tests := []struct {
		name    string
		args    args
		want    *YAML
		wantErr bool
	}{
		{
			name: "Simple append of tasks",
			args: args{
				left: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Command: "left postrollout 1"}},
							{Run: Task{Command: "left postrollout 2"}},
						}},
				},
				right: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Command: "right postrollout 1"}},
						}},
				},
			},
			want: &YAML{
				Tasks: Tasks{
					Postrollout: []TaskRun{
						{Run: Task{Command: "left postrollout 1"}},
						{Run: Task{Command: "left postrollout 2"}},
						{Run: Task{Command: "right postrollout 1"}},
					},
				},
			},
		},
		{
			name: "Merging tasks with the same name",
			args: args{
				left: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Name: "Override me", Command: "left postrollout 1", Container: "should not be overwritten"}},
							{Run: Task{Command: "left postrollout 2"}},
						}},
				},
				right: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Name: "Override me", Command: "right postrollout 1"}},
						}},
				},
			},
			want: &YAML{
				Tasks: Tasks{
					Postrollout: []TaskRun{
						{Run: Task{Name: "Override me", Command: "right postrollout 1", Container: "should not be overwritten"}},
						{Run: Task{Command: "left postrollout 2"}},
					},
				},
			},
		},
		{
			name: "Merging tasks with weight",
			args: args{
				left: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Command: "left postrollout 1", Weight: 0}},
							{Run: Task{Command: "left postrollout 2", Weight: 0}},
						}},
				},
				right: &YAML{
					Tasks: Tasks{
						Postrollout: []TaskRun{
							{Run: Task{Command: "Right comes before", Weight: -1}},
							{Run: Task{Command: "Right comes after", Weight: 1}},
						}},
				},
			},
			want: &YAML{
				Tasks: Tasks{
					Postrollout: []TaskRun{
						{Run: Task{Command: "Right comes before", Weight: -1}},
						{Run: Task{Command: "left postrollout 1", Weight: 0}},
						{Run: Task{Command: "left postrollout 2", Weight: 0}},
						{Run: Task{Command: "Right comes after", Weight: 1}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//fmt.Println(tt.args.left)
			//fmt.Println(tt.args.right)
			err := MergeLagoonYAMLs(tt.args.left, tt.args.right)
			got := tt.args.left
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeLagoonYAMLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeLagoonYAMLs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

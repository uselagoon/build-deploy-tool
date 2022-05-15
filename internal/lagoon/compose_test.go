package lagoon

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshaDockerComposeYAML(t *testing.T) {
	type args struct {
		file string
		l    *Compose
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *Compose
	}{
		{
			name: "test1 docker-compose drupal example",
			args: args{
				file: "test-resources/drupal-docker-compose.yml",
				l:    &Compose{},
			},
			want: &Compose{
				Services: map[string]Service{
					"cli": {
						Build: ServiceBuild{
							Context:    ".",
							Dockerfile: "lagoon/cli.dockerfile",
						},
						Labels: map[string]string{
							"lando.type":             "php-cli-drupal",
							"lagoon.persistent":      "/app/web/sites/default/files/",
							"lagoon.persistent.name": "nginx",
							"lagoon.type":            "cli-persistent",
						},
					},
					"nginx": {
						Build: ServiceBuild{
							Context:    ".",
							Dockerfile: "lagoon/nginx.dockerfile",
						},
						Labels: map[string]string{
							"lando.type":        "nginx-drupal",
							"lagoon.persistent": "/app/web/sites/default/files/",
							"lagoon.type":       "nginx-php-persistent",
						},
					},
					"php": {
						Build: ServiceBuild{
							Context:    ".",
							Dockerfile: "lagoon/php.dockerfile",
						},
						Labels: map[string]string{
							"lando.type":        "php-fpm",
							"lagoon.persistent": "/app/web/sites/default/files/",
							"lagoon.name":       "nginx",
							"lagoon.type":       "nginx-php-persistent",
						},
					},
					"mariadb": {
						Labels: map[string]string{
							"lando.type":  "mariadb-drupal",
							"lagoon.type": "mariadb",
						},
					},
					"redis": {
						Labels: map[string]string{
							"lando.type":  "redis",
							"lagoon.type": "redis",
						},
					},
					"solr": {
						Labels: map[string]string{
							"lando.type":  "solr-drupal",
							"lagoon.type": "solr",
						},
					},
				},
			},
		},
		{
			name: "test2 docker-compose node example",
			args: args{
				file: "test-resources/node-docker-compose.yml",
				l:    &Compose{},
			},
			want: &Compose{
				Services: map[string]Service{
					"node": {
						Build: ServiceBuild{
							Context:    ".",
							Dockerfile: "node.dockerfile",
						},
						Labels: map[string]string{
							"lagoon.type": "node",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UnmarshaDockerComposeYAML(tt.args.file, tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshaDockerComposeYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.l, tt.want) {
				stra, _ := json.Marshal(tt.args.l)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("UnmarshaDockerComposeYAML() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}

func TestCheckLagoonLabel(t *testing.T) {
	type args struct {
		labels map[string]string
		label  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				labels: map[string]string{
					"lagoon.type":            "cli-persistent",
					"lagoon.persistent":      "/app/web/sites/default/files/",
					"lagoon.persistent.name": "nginx",
				},
				label: "lagoon.persistent.name",
			},
			want: "nginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckServiceLagoonLabel(tt.args.labels, tt.args.label); got != tt.want {
				t.Errorf("CheckServiceLagoonLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

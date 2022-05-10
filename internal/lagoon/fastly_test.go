package lagoon

import (
	"reflect"
	"testing"
)

func TestGenerateFastlyConfiguration(t *testing.T) {
	type args struct {
		noCacheServiceID string
		serviceID        string
		route            string
		secretPrefix string
		variables        []EnvironmentVariable
	}
	tests := []struct {
		name    string
		args    args
		want    Fastly
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				noCacheServiceID: "",
				serviceID:        "",
				route:            "",
				secretPrefix: "",
				variables: []EnvironmentVariable{
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567:true",
						Scope: "global",
					},
				},
			},
			want: Fastly{
				Watch:     true,
				ServiceID: "1234567",
			},
		},
		{
			name: "test2",
			args: args{
				noCacheServiceID: "",
				serviceID:        "",
				route:            "",
				secretPrefix: "",
				variables: []EnvironmentVariable{
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567:true:secretname",
						Scope: "global",
					},
				},
			},
			want: Fastly{
				Watch:         true,
				ServiceID:     "1234567",
				APISecretName: "secretname",
			},
		},
		{
			name: "test3",
			args: args{
				noCacheServiceID: "",
				serviceID:        "",
				route:            "www.example.com",
				secretPrefix: "api-secret-",
				variables: []EnvironmentVariable{
					{
						Name:  "LAGOON_FASTLY_SERVICE_IDS",
						Value: "www.example.com:abcdefg:true:secretname,example.com:1234567:true:secretname",
						Scope: "global",
					},
				},
			},
			want: Fastly{
				Watch:         true,
				ServiceID:     "abcdefg",
				APISecretName: "api-secret-secretname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFastlyConfiguration(tt.args.noCacheServiceID, tt.args.serviceID, tt.args.route, tt.args.secretPrefix, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateFastlyAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateFastlyAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

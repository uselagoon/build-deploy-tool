package lagoon

import (
	"fmt"
	"reflect"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestTaskUnmarshall(t *testing.T) {
	lagoonYmlTasks := "tasks:\n  pre-rollout:\n    - run:\n        name: drush sql-dump\n        command: mkdir -p /app/web/sites/default/files/private/ && drush sql-dump --ordered-dump --gzip --result-file=/app/web/sites/default/files/private/pre-deploy-dump.sql.gz\n        service: cli\n  post-rollout:\n    - run:\n        name: drush cim\n        command: drush -y cim\n        service: cli\n        shell: bash\n    - run:\n        name: drush cr\n        command: drush -y cr\n        service: cli"
	//lagoonYmlTasks := "        name: drush cim\n        command: drush -y cim\n        service: cli\n        shell: bash"
	var lYAML YAML
	//var lYAML Task
	yaml.Unmarshal([]byte(lagoonYmlTasks), &lYAML)
	fmt.Println(lYAML)
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name string
		want Task
	}{
		{
			name: "Test empty new Task",
			want: Task{
				Command:   "",
				Namespace: "",
				Service:   "cli",
				Shell:     "sh",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTask(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

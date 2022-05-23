package lagoon

import (
	"reflect"
	"testing"
)

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

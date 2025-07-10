package cron

import (
	"testing"
	"sort"
	"reflect"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

func ptr(b bool) *bool {
	return &b
}

func TestStandardizeSchedule(t *testing.T) {
	tests := []struct {
		name       string
		cron       string
		want       string
		wantErrMsg string
		wantErr    bool
	}{
		{
			name: "test1",
			cron: "M * * * *",
			want: "42 * * * *",
		},
		{
			name: "test2",
			cron: "M/5 * * * *",
			want: "2,7,12,17,22,27,32,37,42,47,52,57 * * * *",
		},
		{
			name: "test3",
			cron: "M H(2-4) * * *",
			want: "42 2 * * *",
		},
		{
			name: "test4",
			cron: "M H(22-2) * * *",
			want: "42 0 * * *",
		},
		{
			name: "test5",
			cron: "M/15 H(22-2) * * *",
			want: "12,27,42,57 0 * * *",
		},
		{
			name: "test8",
			cron: "M/15 H(22-2) 3,5 * *",
			want: "12,27,42,57 0 3,5 * *",
		},
		{
			name: "test9",
			cron: "M/15 H(22-2) * 10-12 *",
			want: "12,27,42,57 0 * 10-12 *",
		},
		{
			name: "test14 - set hours",
			cron: "M/15 23 * * 0-5",
			want: "12,27,42,57 23 * * 0-5",
		},
		{
			name: "test15 - set day",
			cron: "M/15 * 31 * 0-5",
			want: "12,27,42,57 * 31 * 0-5",
		},
		{
			name: "test16 - set month",
			cron: "M/15 * * 11 0-5",
			want: "12,27,42,57 * * 11 0-5",
		},
		{
			name: "test17 - hourly interval",
			cron: "M */6 * * *",
			want: "42 0,6,12,18 * * *",
		},
		{
			name: "test18 - day and month string",
			cron: "M */6 * JAN MON",
			want: "42 0,6,12,18 * JAN MON",
		},
		{
			name: "test19 - whitespace",
			cron: "M * * * * ",
			want: "42 * * * *",
		},
		{
			name: "test22 - split range hours",
			cron: "*/30 0-12,22-23 * * *",
			want: "12,42 0-12,22-23 * * *",
		},
		{
			name: "test25 - random minute with step and range hours",
			cron: "M/30 17/12,0-23 * * *",
			want: "12,42 17/12,0-23 * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cj := &Cronjob{
				Schedule: tt.cron,
			}

			schedule, err := StandardizeSchedule(cj.Schedule, 42)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %v, wantErrMsg = %v", err.Error(), tt.wantErrMsg)
				}
			} else {
				if tt.wantErr {
					t.Errorf("expected error, got none")
				}
				if schedule != tt.want {
					t.Errorf("Schedule = %v, want = %v", schedule, tt.want)
				}
			}
		})
	}
}


func TestNormalizeSchedule(t *testing.T) {
	tests := []struct {
		name           string
		input          *CronSchedule
		expectedMinute []int
		expectedHour   []int
		expectError    bool
	}{
		{
			name: "valid multiple values",
			input: &CronSchedule{
				Minute: "0,15,30,45",
				Hour:   "0,6,12,18",
				Day:    "*",
				Month:  "*",
				DayOfWeek: "*",
			},
			expectedMinute: []int{0, 15, 30, 45},
			expectedHour:   []int{0, 6, 12, 18},
			expectError:    false,
		},
		{
			name: "valid range",
			input: &CronSchedule{
				Minute: "0-10",
				Hour:   "0",
				Day:    "*",
				Month:  "*",
				DayOfWeek: "*",
			},
			expectedMinute: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectedHour:   []int{0},
			expectError:    false,
		},
		{
			name: "valid star increment",
			input: &CronSchedule{
				Minute: "*/15",
				Hour:   "0",
				Day:    "*",
				Month:  "*",
				DayOfWeek: "*",
			},
			expectedMinute: []int{0, 15, 30, 45},
			expectedHour:   []int{0},
			expectError:    false,
		},
		{
			name: "invalid minute range",
			input: &CronSchedule{
				Minute: "0,100", // 100 invalid minute
				Hour:   "0",
				Day:    "*",
				Month:  "*",
				DayOfWeek: "*",
			},
			expectError: true,
		},
		{
			name: "empty schedule fields",
			input: &CronSchedule{
				Minute: "",
				Hour:   "",
				Day:    "*",
				Month:  "*",
				DayOfWeek: "*",
			},
			expectedMinute: []int{},
			expectedHour:   []int{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := normalizeSchedule(tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				return // error expected, test done
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotMinutes := schedule[MINUTE_INDEX].ToSlice()
			gotHours := schedule[HOUR_INDEX].ToSlice()

			sort.Ints(gotMinutes)
			sort.Ints(gotHours)
			sort.Ints(tt.expectedMinute)
			sort.Ints(tt.expectedHour)

			if !reflect.DeepEqual(gotMinutes, tt.expectedMinute) {
				t.Errorf("minutes: expected %v, got %v", tt.expectedMinute, gotMinutes)
			}
			if !reflect.DeepEqual(gotHours, tt.expectedHour) {
				t.Errorf("hours: expected %v, got %v", tt.expectedHour, gotHours)
			}
		})
	}
}

func TestFlattenSchedule(t *testing.T) {
	makeSet := func(nums ...int) mapset.Set[int] {
		s := mapset.NewSet[int]()
		s.Append(nums...)
		return s
	}

	tests := []struct {
		name        string
		schedule    []mapset.Set[int]
		expected    []int 
		expectError bool
	}{
		{
			name: "valid schedule with 2 minutes and 2 hours",
			schedule: []mapset.Set[int]{
				makeSet(0, 30),        // minutes
				makeSet(1, 2),         // hours
				makeSet(), makeSet(), makeSet(), // ignored fields
			},
			expected: []int{60, 90, 120, 150},
		},
		{
			name: "valid schedule with some hours set",
			schedule: []mapset.Set[int]{
				makeSet(0),        // minutes
				makeSet(0, 6, 12, 18),         // hours
				makeSet(), makeSet(), makeSet(), // ignored fields
			},
			expected: []int{0, 360, 720, 1080}, 
		},
		{
			name: "invalid schedule length",
			schedule: []mapset.Set[int]{
				makeSet(0), 
			},
			expectError: true,
		},
		{
			name: "empty sets still valid",
			schedule: []mapset.Set[int]{
				makeSet(), makeSet(), makeSet(), makeSet(), makeSet(),
			},
			expected: []int{}, // no minutes or hours
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSet, err := flattenSchedule(tt.schedule)
			if (err != nil) != tt.expectError {
				t.Fatalf("[%s] expected error=%v, got error=%v", tt.name, tt.expectError, err)
			}
			if err != nil {
				return
			}

			got := gotSet.ToSlice()
			sort.Ints(got)
			sort.Ints(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("[%s] expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}


func TestDecideRunner(t *testing.T) {
	tests := []struct {
		name           string
		cronjob        Cronjob
		expectedInPod  *bool
		expectError    bool
	}{
		{
			name: "InPod is already set - skip logic",
			cronjob: Cronjob{
				InPod: ptr(true), // assume ptr returns *bool
			},
			expectedInPod: ptr(true),
		},
		{
			name: "metric below ceiling - should set InPod true",
			cronjob: Cronjob{
				Schedule: "0,15,30,45 * * * *", // well-spaced
			},
			expectedInPod: ptr(true),
		},
		{
			name: "metric above ceiling - should set InPod false",
			cronjob: Cronjob{
				Schedule: "0 0,6,12,18 * * *", // widely spaced
			},
			expectedInPod: ptr(false),
		},
		{
			name: "invalid schedule - returns error",
			cronjob: Cronjob{
				Schedule: "invalid-schedule",
			},
			expectError: true,
		},
		{
			name: "InPod=true test 2-59/10 * * * *",
			cronjob: Cronjob{
				Schedule: "2-59/10 * * * *", // well-spaced
			},
			expectedInPod: ptr(true),
		},
		{
			name: "InPod=true test 1-59/10 * * * *",
			cronjob: Cronjob{
				Schedule: "1-59/10 * * * *", // well-spaced
			},
			expectedInPod: ptr(true),
		},
		{
			name: "InPod=true test 3-59/5 * * * *",
			cronjob: Cronjob{
				Schedule: "3-59/5 * * * *", // well-spaced
			},
			expectedInPod: ptr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := tt.cronjob.decideRunner()

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.cronjob.InPod == nil {
				t.Fatalf("InPod is nil, expected %v", *tt.expectedInPod)
			}
			if *tt.cronjob.InPod != *tt.expectedInPod {
				t.Errorf("expected InPod = %v, got %v, metric was %d, schedule was %s", *tt.expectedInPod, *tt.cronjob.InPod, metric, tt.cronjob.Schedule)
			}
		})
	}
}

func TestCalculateMetric(t *testing.T) {
	tests := []struct {
		name     string
		input    []int           // Times in minutes
		expected int             // Expected median distance
		wantErr  bool
	}{
		{
			name:     "uniform spacing",
			input:    []int{0, 60, 120, 180}, // distances: 60, 60, 60, 1260 -> median: 60
			expected: 60,
		},
		{
			name:     "non-uniform spacing",
			input:    []int{0, 15, 45, 1200}, // distances: 15, 30, 1155, 240 -> median: 135
			expected: 135,
		},
		{
			name:     "one element",
			input:    []int{300}, // median of a single element is itself
			expected: 300,
		},
		{
			name:    "empty set",
			input:   []int{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := mapset.NewSet[int]()
			set.Append(tt.input...)

			got, err := calculateMetric(set)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err != nil)
			}

			if !tt.wantErr && got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

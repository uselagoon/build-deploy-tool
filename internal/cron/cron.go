package cron

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	cron "github.com/robfig/cron/v3"
	"github.com/cxmcc/unixsums/cksum"
)

// Cronjob represents a Lagoon cronjob.
type Cronjob struct {
	Name     string `json:"name"`
	Service  string `json:"service"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	InPod    *bool  `json:"inPod"`
	Timeout  string `json:"timeout"`
}

type CronSchedule struct {
	Minute    string
	Hour      string
	Day       string
	Month     string
	DayOfWeek string
}

const SCHEDULE_METRIC_CEILING = 30

func (c *CronSchedule) String() string {
	return fmt.Sprintf("%s %s %s %s %s", c.Minute, c.Hour, c.Day, c.Month, c.DayOfWeek)
}

// this will check if someone has requested a pseudo random interval for the minute using `M/` or `H/`
func (c *CronSchedule) validateReplaceMinute(seed uint32) error {
	match1, _ := regexp.MatchString("^(M|H)$", c.Minute)
	if match1 {
		// If just an `M` or `H` (for backwards compatibility) is defined, we
		// generate a pseudo random minute.
		c.Minute = strconv.Itoa(int(math.Mod(float64(seed), 60)))
	}
	match2, _ := regexp.MatchString("^(M|H|\\*)/([0-5]?[0-9])$", c.Minute)
	if match2 {
		// A Minute like M/15 (or H/15 or */15 for backwards compatibility) is defined, create a list of minutes with a random start
		// like 4,19,34,49 or 6,21,36,51
		params := getCaptureBlocks("^(?P<P1>M|H|\\*)/(?P<P2>[0-5]?[0-9])$", c.Minute)
		step, err := strconv.Atoi(params["P2"])
		if err != nil {
			return fmt.Errorf("unable to determine hours value")
		}
		counter := int(math.Mod(float64(seed), float64(step)))
		var minutesArr []string
		for counter < 60 {
			minutesArr = append(minutesArr, fmt.Sprintf("%d", counter))
			counter += step
		}
		c.Minute = strings.Join(minutesArr, ",")
	}
	return nil
}

// this will check if someone has requested a pseudo random interval for the hour using `H/`
func (c *CronSchedule) validateReplaceHour(seed uint32) error {
	match1, _ := regexp.MatchString("^H$", c.Hour)
	if match1 {
		// If just an `H` is defined, we generate a pseudo random hour.
		c.Hour = strconv.Itoa(int(math.Mod(float64(seed), 24)))
	}
	match2, _ := regexp.MatchString("^H\\(([01]?[0-9]|2[0-3])-([01]?[0-9]|2[0-3])\\)$", c.Hour)
	if match2 {
		// If H is defined with a given range, example: H(2-4), we generate a random hour between 2-4
		params := getCaptureBlocks("^H\\((?P<P1>[01]?[0-9]|2[0-3])-(?P<P2>[01]?[0-9]|2[0-3])\\)$", c.Hour)
		hFrom, err := strconv.Atoi(params["P1"])
		if err != nil {
			return fmt.Errorf("unable to determine hours value")
		}
		hTo, err := strconv.Atoi(params["P2"])
		if err != nil {
			return fmt.Errorf("unable to determine hours value")
		}
		if hFrom < hTo {
			// Example: HOUR_FROM: 2, HOUR_TO: 4
			// Calculate the difference between the two hours (in example will be 2)
			maxDiff := float64(hTo - hFrom)
			// Generate a difference based on the SEED (in example will be 0, 1 or 2)
			diff := int(math.Mod(float64(seed), maxDiff))
			// Add the generated difference to the FROM hour (in example will be 2, 3 or 4)
			c.Hour = strconv.Itoa(hFrom + diff)
			return nil
		}
		if hFrom > hTo {
			// If the FROM is larger than the TO, we have a range like 22-2
			// Calculate the difference between the two hours with a 24 hour jump (in example will be 4)
			maxDiff := float64(24 - hFrom + hTo)
			// Generate a difference based on the SEED (in example will be 0, 1, 2, 3 or 4)
			diff := int(math.Mod(float64(seed), maxDiff))
			// Add the generated difference to the FROM hour (in example will be 22, 23, 24, 25 or 26)
			if hFrom+diff >= 24 {
				// If the hour is higher than 24, we subtract 24 to handle the midnight change
				c.Hour = strconv.Itoa(hFrom + diff - 24)
				return nil
			}
			c.Hour = strconv.Itoa(hFrom + diff)
			return nil
		}
		if hFrom == hTo {
			c.Hour = strconv.Itoa(hFrom)
		}
	}
	match3, _ := regexp.MatchString("^(H|\\*)/([01]?[0-9]|2[0-3])$", c.Hour)
	if match3 {
		// An hour like H/15 or */15 is defined, create a list of hours with a random start
		// like 1,7,13,19
		params := getCaptureBlocks("^(?P<P1>H|\\*)/(?P<P2>[01]?[0-9]|2[0-3])$", c.Hour)
		step, err := strconv.Atoi(params["P2"])
		if err != nil {
			return fmt.Errorf("unable to determine hours value")
		}
		counter := int(math.Mod(float64(seed), float64(step)))
		var hoursArr []string
		for counter < 24 {
			hoursArr = append(hoursArr, fmt.Sprintf("%d", counter))
			counter += step
		}
		c.Hour = strings.Join(hoursArr, ",")
	}
	return nil
}

func getCaptureBlocks(regex, val string) (captureMap map[string]string) {
	var regexComp = regexp.MustCompile(regex)
	match := regexComp.FindStringSubmatch(val)
	captureMap = make(map[string]string)
	for i, name := range regexComp.SubexpNames() {
		if i > 0 && i <= len(match) {
			captureMap[name] = match[i]
		}
	}
	return captureMap
}

// Process lagoon cronjob to ensure it meets our constraints.
func (cj *Cronjob) ValidateCronjob() error {

	// Use checksum of cron command as the seed
	seed := cksum.Cksum([]byte(fmt.Sprintf("%s\n", cj.Command)))
	schedule, err := StandardizeSchedule(cj.Schedule, seed)
	if err != nil {
		return err
	}
	cj.Schedule = schedule

	_, err = cj.decideRunner()
	if err != nil {
		return err
	}

	err = cj.validateTimeout()
	if err != nil {
		return err
	}

	return nil
}

// Convert and validate a cron schedule to the stanardized format.
// In particular, this means derandomizing, i.e removing the "H" and "M" that may be present.
// We consume the seed so that we can enforce determinism for testing purposes.
func StandardizeSchedule(schedule string, seed uint32) (string, error) {
	splitSchedule := strings.Split(strings.Trim(schedule, " "), " ")

	if len(splitSchedule) <= 0 {
		err := fmt.Errorf("cron definition '%s' is invalid", schedule)
		schedule = ""
		return schedule, err
	}

	if len(splitSchedule) != 5 {
		err := fmt.Errorf("cron definition '%s' is invalid, %d fields provided, required 5", schedule, len(splitSchedule))
		schedule = ""
		return schedule, err
	}

	newSchedule := &CronSchedule{
		Minute:    splitSchedule[0],
		Hour:      splitSchedule[1],
		Day:       splitSchedule[2],
		Month:     splitSchedule[3],
		DayOfWeek: splitSchedule[4],
	}
	// validate for any M/H style replacements for pseudo-random intervals
	if err := newSchedule.validateReplaceMinute(seed); err != nil {
		err = fmt.Errorf("cron definition '%s' is invalid, unable to determine minutes value", schedule)
		schedule = ""
		return schedule, err
	}
	// validate for any H style replacements for pseudo-random intervals
	if err := newSchedule.validateReplaceHour(seed); err != nil {
		err = fmt.Errorf("cron definition '%s' is invalid, unable to determine hours value", schedule)
		schedule = ""
		return schedule, err
	}
	// parse/validate the same as kubernetes once the pseudo-random intervals have been calculated and updated into the schedule
	// https://github.com/kubernetes/kubernetes/blob/58c44005cdaec53fe3cb49b2d7a308df3af2d081/pkg/controller/cronjob/cronjob_controllerv2.go#L394
	if _, err := cron.ParseStandard(newSchedule.String()); err != nil {
		err = fmt.Errorf("cron definition '%s' is invalid", schedule)
		schedule = ""
		return schedule, err
	}

	schedule = newSchedule.String()
	return schedule, nil
}

// Decide, based on the frequency of jobs defined in the cron schedule,
// whether the cronjob should run in a pod or as a k8s-native job.
// This function should never be called before the schedule is stanardized,
// as the logic does not handle randomized scheduling.
func (cj *Cronjob) decideRunner() (int, error) {
	if cj.InPod != nil {
		// Don't perform decision algorithm if InPod has been explicitly set.
		return 0, nil
	}

	metric, err := calculateScheduleMetric(cj.Schedule)
	if err != nil {
		return 0, err
	}

	inPod := (metric <= SCHEDULE_METRIC_CEILING)
	cj.InPod = &inPod

	return metric, nil
}

// Validate the cronjob's timeout string is valid. It can't be greater than 24hrs
// and must match go time duration https://pkg.go.dev/time#ParseDuration
func (cj *Cronjob) validateTimeout() error {
	if cj.Timeout == "" {
		// default cronjob timeout is 4h
		cj.Timeout = "4h"
		return nil
	}

	cjTimeout, err := time.ParseDuration(cj.Timeout)
	if err != nil {
		return fmt.Errorf("unable to convert timeout for cronjob %s: %v", cj.Name, err)
	}
	// max cronjob timeout is 24 hours
	if cjTimeout > time.Duration(24*time.Hour) {
		return fmt.Errorf("timeout for cronjob %s cannot be longer than 24 hours", cj.Name)
	}

	return nil
}

func calculateScheduleMetric(schedule string) (int, error) {
	if schedule == "" {
		return 0, fmt.Errorf("Schedule cant be empty")
	}

	splitSchedule := strings.Split(strings.Trim(schedule, " "), " ")
	if len(splitSchedule) != 5 {
		return 0, fmt.Errorf("Bad schedule string %s", schedule)
	}

	cs := &CronSchedule{
		Minute:    splitSchedule[0],
		Hour:      splitSchedule[1],
		Day:       splitSchedule[2],
		Month:     splitSchedule[3],
		DayOfWeek: splitSchedule[4],
	}

	timeSets, err := normalizeSchedule(cs)
	if err != nil {
		return 0, err
	}

	flatSchedule, err := flattenSchedule(timeSets)
	if err != nil {
		return 0, err
	}

	metric, err := calculateMetric(flatSchedule)
	if err != nil {
		return 0, err
	}

	return metric, nil
}

// DOM := Day of Month
// DOW := Day of Week
const (
	MINUTE_INDEX = 0
	HOUR_INDEX   = 1
	DOM_INDEX    = 2
	MONTH_INDEX  = 3
	DOW_INDEX    = 4
)

// Converts each field in the schedule from the concise
// a1[-b1][/x1],a2[-b2][/x2],a3[-b3][/x3],...
// to the 'normalized' y1,y2,y3,... which makes further processing much easier.
func normalizeSchedule(cs *CronSchedule) ([]mapset.Set[int], error) {
	sets := make([]mapset.Set[int], 5)
	var err error

	for i := range sets {
		sets[i] = mapset.NewSet[int]()
	}

	sets[MINUTE_INDEX], err = normalizeField(cs.Minute, 0, 59)
	if err != nil {
		return nil, err
	}
	sets[HOUR_INDEX], err = normalizeField(cs.Hour, 0, 23)
	if err != nil {
		return nil, err
	}

	return sets, nil
}

// Converts a string of form "a1[-b1][/x1],a2[-b2][/x2],a3[-b3][/x3],..."
// (square brackets indicating optionality) into the normalized
// y1,y2,y3,... where each y is bounded by the max and min
// allowed values of that field. For example, minutes can be 0 to 59, while DOM
// is from 1 to 31.
func normalizeField(field string, min int, max int) (mapset.Set[int], error) {
	if field == "" {
		return nil, fmt.Errorf("field string cannot be empty")
	}

	var set = mapset.NewSet[int]()
	// All pattern 
	allPattern := regexp.MustCompile(`^\*$`)
	// Increment pattern
	stepOnlyPattern := regexp.MustCompile(`^\*/(\d{1,2})$`)
	// Matches a single time
	singleTimePattern := regexp.MustCompile(`^\d{1,2}$`)
	// Matches a range of times
	rangePattern := regexp.MustCompile(`^(\d{1,2})-(\d{1,2})$`)
	// Matches a range of times with increment
	rangeWithIncPattern := regexp.MustCompile(`^(\d{1,2})-(\d{1,2})/(\d{1,2})$`)

	timeRanges := strings.Split(field, ",")
	for _, timeRange := range timeRanges {
		if allPattern.MatchString(timeRange) {
			for i := min; i <= max; i++ {
				set.Add(i)
			}
		} else if stepOnlyPattern.MatchString(timeRange) {
			step, err := strconv.Atoi(stepOnlyPattern.FindStringSubmatch(timeRange)[1])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step value in %s", timeRange)
			}
			for i := min; i <= max; i += step {
				set.Add(i)
			}
		} else if singleTimePattern.MatchString(timeRange) {
			time, err := strconv.Atoi(timeRange)
			if err != nil {
				return nil, err
			}
			set.Add(time)
		} else if rangePattern.MatchString(timeRange) {
			parts := rangePattern.FindStringSubmatch(timeRange)
			start, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}

			end, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}

			if end < start || start < min || end > max {
				return nil, fmt.Errorf("Invalid range %s", timeRange)
			}

			for i := start; i <= end; i++ {
				set.Add(i)
			}
		} else if rangeWithIncPattern.MatchString(timeRange) {
			parts := rangeWithIncPattern.FindStringSubmatch(timeRange)
			// The range with increment pattern has 3 capture groups,
			// so parts should definitely be of size 4.
			if len(parts) != 4 {
				return nil, fmt.Errorf("Broken Invariant: wrong number of capture groups (%d) returned from %s", len(parts), timeRange)
			}

			start, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}

			end, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}

			inc, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, err
			}

			if end < start || inc <= 0 || inc > 60 {
				return nil, fmt.Errorf("Invalid range %s", timeRange)
			}

			for i := start; i <= end; i += inc {
				set.Add(i)
			}
		} else {
			return nil, fmt.Errorf("Invalid range in schedule %s, range %s", field, timeRange)
		}
	}

	return set, nil
}

// Converts a normalized cron schedule from 5 sets of times to a
// set of times from 0 to 1440 (the number of minutes in a day).
// We do this because by definition, a cron schedule defines a set of integers in this range.
func flattenSchedule(schedule []mapset.Set[int]) (mapset.Set[int], error) {
	var flat = mapset.NewSet[int]()

	if len(schedule) != 5 {
		return nil, fmt.Errorf("Schedule of wrong size passed in. Should be 5, was %d", len(schedule))
	}

	for _, minute := range schedule[0].ToSlice() {
		for _, hour := range schedule[1].ToSlice() {
			flat.Add(hour*60 + minute)
		}
	}

	return flat, nil
}

func calculateMetric(times mapset.Set[int]) (int, error) {
	if times == nil || times.Cardinality() == 0 {
		return 0, fmt.Errorf("Cannot calculateMetric on empty set")
	}

	if times.Cardinality() == 1 {
		return times.ToSlice()[0], nil
	}

	// Sort the times
	values := times.ToSlice()
	sort.Ints(values)

	var distances []int
	for i := 1; i < len(values); i++ {
		diff := values[i] - values[i-1]
		distances = append(distances, diff)
	}

	// Measure distance between last time in a day and the start of the next day
	circularDiff := (values[0] + 1440) - values[len(values)-1]
	distances = append(distances, circularDiff)

	// use median metric
	metric := calculateMedian(distances)

	return metric, nil
}

func calculateMedian(values []int) int {
	sort.Ints(values)

	n := len(values)
	mid := n / 2

	if n%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	} else {
		return values[mid]
	}
}

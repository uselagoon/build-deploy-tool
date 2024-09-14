package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/cxmcc/unixsums/cksum"

	cron "github.com/robfig/cron/v3"
)

type Cron struct {
	Minute    string
	Hour      string
	Day       string
	Month     string
	DayOfWeek string
}

func (c *Cron) String() string {
	return fmt.Sprintf("%s %s %s %s %s", c.Minute, c.Hour, c.Day, c.Month, c.DayOfWeek)
}

// this will check if someone has requested a pseudo random interval for the minute using `M/` or `H/`
func (c *Cron) validateReplaceMinute(seed uint32) error {
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
func (c *Cron) validateReplaceHour(seed uint32) error {
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

func ConvertCrontab(namespace, schedule string) (string, error) {
	splitSchedule := strings.Split(strings.Trim(schedule, " "), " ")
	seed := cksum.Cksum([]byte(fmt.Sprintf("%s\n", namespace)))
	if len(splitSchedule) == 5 {
		newSchedule := &Cron{
			Minute:    splitSchedule[0],
			Hour:      splitSchedule[1],
			Day:       splitSchedule[2],
			Month:     splitSchedule[3],
			DayOfWeek: splitSchedule[4],
		}
		// validate for any M/H style replacements for pseudo-random intervals
		if err := newSchedule.validateReplaceMinute(seed); err != nil {
			return "", fmt.Errorf("cron definition '%s' is invalid, unable to determine minutes value", schedule)
		}
		// validate for any H style replacements for pseudo-random intervals
		if err := newSchedule.validateReplaceHour(seed); err != nil {
			return "", fmt.Errorf("cron definition '%s' is invalid, unable to determine hours value", schedule)
		}
		// parse/validate the same as kubernetes once the pseudo-random intervals have been calculated and updated into the schedule
		// https://github.com/kubernetes/kubernetes/blob/58c44005cdaec53fe3cb49b2d7a308df3af2d081/pkg/controller/cronjob/cronjob_controllerv2.go#L394
		if _, err := cron.ParseStandard(newSchedule.String()); err != nil {
			return "", fmt.Errorf("cron definition '%s' is invalid", schedule)
		}
		// if valid, return the cron schedule string value
		return newSchedule.String(), nil
	}
	if len(splitSchedule) < 5 && len(splitSchedule) > 0 || len(splitSchedule) > 5 {
		return "", fmt.Errorf("cron definition '%s' is invalid, %d fields provided, required 5", schedule, len(splitSchedule))
	}
	return "", fmt.Errorf("cron definition '%s' is invalid", schedule)
}

func IsInPodCronjob(schedule string) (bool, error) {
	splitSchedule := strings.Split(strings.Trim(schedule, " "), " ")
	// check the provided cron splits into 5
	if len(splitSchedule) == 5 {
		for idx, val := range splitSchedule {
			if idx == 0 {
				match1, _ := regexp.MatchString("^(M|H|\\*)/([0-5]?[0-9])$", val)
				if match1 {
					// A Minute like M/15 (or H/15 or */15 for backwards compatibility) is defined, create a list of minutes with a random start
					// like 4,19,34,49 or 6,21,36,51
					params := getCaptureBlocks("^(?P<P1>M|H|\\*)/(?P<P2>[0-5]?[0-9])$", val)
					step, err := strconv.Atoi(params["P2"])
					if err != nil {
						return false, fmt.Errorf("cron definition '%s' is invalid, unable to determine minutes value", schedule)
					}
					if step < 30 {
						return true, nil
					}
				}
				match2, _ := regexp.MatchString("^\\*$", val)
				if match2 {
					// this runs every minute
					return true, nil
				}
			}
		}
	}
	return false, nil
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

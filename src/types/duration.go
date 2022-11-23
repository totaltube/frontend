package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(DurationToHuman(time.Duration(d))), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	*d = Duration(ParseHumanDuration(string(text)))
	return nil
}

func (d Duration) String() string {
	return DurationToHuman(time.Duration(d))
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(DurationToHuman(time.Duration(d)))
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		if !regexp.MustCompile(`^\d+h\d+m\d+s$`).MatchString(value) {
			*d = Duration(ParseHumanDuration(value))
			return nil
		}
		var err error
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func DurationToHuman(d time.Duration) (human string) {
	var res []string
	years := math.Floor(d.Hours() / 8760.0)
	if years >= 1 {
		d = d - time.Duration(years)*8760.0*time.Hour
		if years == 1 {
			res = append(res, fmt.Sprintf("%d year", int64(years)))
		} else {
			res = append(res, fmt.Sprintf("%d years", int64(years)))
		}
	}
	months := math.Floor(d.Hours() / 730.0)
	if months >= 1 {
		d = d - time.Duration(months)*730.0*time.Hour
		if months == 1 {
			res = append(res, fmt.Sprintf("%d month", int64(months)))
		} else {
			res = append(res, fmt.Sprintf("%d months", int64(months)))
		}
	}
	days := math.Floor(d.Hours() / 24)
	if days >= 1 {
		d = d - time.Duration(days)*24*time.Hour
		if days == 1 {
			res = append(res, fmt.Sprintf("%d day", int64(days)))
		} else {
			res = append(res, fmt.Sprintf("%d days", int64(days)))
		}
	}
	hours := math.Floor(d.Hours())
	if hours >= 1 {
		d = d - time.Duration(hours)*time.Hour
		if hours == 1 {
			res = append(res, fmt.Sprintf("%d hour", int64(hours)))
		} else {
			res = append(res, fmt.Sprintf("%d hours", int64(hours)))
		}
	}
	minutes := math.Floor(d.Minutes())
	if minutes >= 1 {
		d = d - time.Duration(minutes)*time.Minute
		if minutes == 1 {
			res = append(res, fmt.Sprintf("%d minute", int64(minutes)))
		} else {
			res = append(res, fmt.Sprintf("%d minutes", int64(minutes)))
		}
	}
	seconds := math.Floor(d.Seconds())
	if seconds > 0 || len(res) == 0 {
		if seconds == 1 {
			res = append(res, fmt.Sprintf("%d second", int64(seconds)))
		} else {
			res = append(res, fmt.Sprintf("%d seconds", int64(seconds)))
		}
	}
	return strings.Join(res, ", ")
}

func ParseHumanDuration(human string) time.Duration {
	var hours float64 = 0
	var minutes float64 = 0
	var seconds float64 = 0
	res := regexp.MustCompile(`([\d.]+)\s*y`).FindStringSubmatch(human)
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			hours += i * 8760
		}
	}
	monthPresent := false
	res = regexp.MustCompile(`([\d.]+)\s*mo`).FindStringSubmatch(human)
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			hours += i * 730
		}
		monthPresent = true
	}
	res = regexp.MustCompile(`([\d.]+)\s*d`).FindStringSubmatch(human)
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			hours += i * 24
		}
	}
	res = regexp.MustCompile(`([\d.]+)\s*h`).FindStringSubmatch(human)
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			hours += i
		}
	}
	if !monthPresent {
		res = regexp.MustCompile(`([\d.]+)\s*m`).FindStringSubmatch(human)
	} else {
		res = regexp.MustCompile(`([\d.]+)\s*mi`).FindStringSubmatch(human)
	}
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			minutes += i
		}
	}
	res = regexp.MustCompile(`(\d+)\s*s`).FindStringSubmatch(human)
	if res != nil {
		if i, err := strconv.ParseFloat(res[1], 64); err == nil {
			seconds += i
		}
	}
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
}

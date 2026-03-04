package api

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" || repeat == " " {
		return "", errors.New("empty repeat")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", errors.New("invalid date format")
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("invalid repeat format")
	}

	switch parts[0] {
	case "y":
		if len(parts) != 1 {
			return "", errors.New("invalid repeat format")
		}
		return nextYearly(date, now)

	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat format")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("invalid repeat format")
		}
		return nextDaily(date, days, now)

	case "w":
		if len(parts) < 2 {
			return "", errors.New("invalid repeat format")
		}
		weekdays, err := parseWeekdays(parts[1])
		if err != nil {
			return "", errors.New("invalid repeat format")
		}
		return nextWeekly(date, weekdays, now)

	case "m":
		if len(parts) < 2 {
			return "", errors.New("invalid repeat format")
		}
		monthdays, months, err := parseMonthly(parts[1:])
		if err != nil {
			return "", errors.New("invalid repeat format")
		}
		return nextMonthly(date, monthdays, months, now)

	default:
		return "", errors.New("unsupported repeat format")
	}
}

func afterNow(date, now time.Time) bool {
	dateStr := date.Format(dateFormat)
	nowStr := now.Format(dateFormat)
	return dateStr > nowStr
}

func nextYearly(date, now time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if date.Month() == 3 && date.Day() == 1 {
			prevDay := date.AddDate(0, 0, -1)
			if prevDay.Month() == 2 && prevDay.Day() == 28 {
				// 29 февраля → 1 марта
			}
		}
		if afterNow(date, now) {
			return date.Format(dateFormat), nil
		}
	}
}

func nextDaily(date time.Time, days int, now time.Time) (string, error) {
	for {
		date = date.AddDate(0, 0, days)
		if afterNow(date, now) {
			return date.Format(dateFormat), nil
		}
	}
}

func nextWeekly(date time.Time, weekdays []int, now time.Time) (string, error) {
	current := now.AddDate(0, 0, 1)
	for i := 0; i < 366; i++ {
		goDay := int(current.Weekday())
		if goDay == 0 {
			goDay = 7
		}
		for _, wd := range weekdays {
			if wd == goDay {
				return current.Format(dateFormat), nil
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return "", errors.New("no matching weekday found")
}

func nextMonthly(date time.Time, monthdays []int, months []int, now time.Time) (string, error) {
	current := now.AddDate(0, 0, 0)
	current = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, time.Local)

	for i := 0; i < 24; i++ {
		if len(months) > 0 {
			match := false
			for _, m := range months {
				if int(current.Month()) == m {
					match = true
					break
				}
			}
			if !match {
				current = current.AddDate(0, 1, 0)
				continue
			}
		}

		lastDay := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()

		for _, md := range monthdays {
			var day int
			if md < 0 {
				day = lastDay + md + 1
			} else {
				day = md
			}

			if day > 0 && day <= lastDay {
				testDate := time.Date(current.Year(), current.Month(), day, 0, 0, 0, 0, time.Local)
				if afterNow(testDate, now) {
					return testDate.Format(dateFormat), nil
				}
			}
		}
		current = current.AddDate(0, 1, 0)
	}
	return "", errors.New("no matching monthday found")
}

func parseWeekdays(s string) ([]int, error) {
	var days []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		d, err := strconv.Atoi(part)
		if err != nil || d < 1 || d > 7 {
			return nil, errors.New("invalid weekday")
		}
		days = append(days, d)
	}
	if len(days) == 0 {
		return nil, errors.New("no weekdays specified")
	}
	return days, nil
}

func parseMonthly(parts []string) ([]int, []int, error) {
	if len(parts) == 0 {
		return nil, nil, errors.New("invalid monthly format")
	}

	monthdays, err := parseMonthDays(parts[0])
	if err != nil {
		return nil, nil, err
	}

	var months []int
	if len(parts) > 1 {
		months, err = parseMonths(parts[1])
		if err != nil {
			return nil, nil, err
		}
	}

	return monthdays, months, nil
}

func parseMonthDays(s string) ([]int, error) {
	var days []int
	s = strings.ReplaceAll(s, " ", ",")
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		d, err := strconv.Atoi(part)
		if err != nil {
			return nil, errors.New("invalid day")
		}
		if d == 0 || d < -31 || d > 31 {
			return nil, errors.New("invalid day value")
		}
		days = append(days, d)
	}
	if len(days) == 0 {
		return nil, errors.New("no days specified")
	}
	return days, nil
}

func parseMonths(s string) ([]int, error) {
	var months []int
	s = strings.ReplaceAll(s, " ", ",")
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		m, err := strconv.Atoi(part)
		if err != nil || m < 1 || m > 12 {
			return nil, errors.New("invalid month")
		}
		months = append(months, m)
	}
	if len(months) == 0 {
		return nil, errors.New("no months specified")
	}
	return months, nil
}

func CalculateNextDate(dateStr, repeatStr string) string {
	result, _ := NextDate(time.Now(), dateStr, repeatStr)
	return result
}

func parseDays(s string) []int {
	var days []int
	s = strings.ReplaceAll(s, " ", ",")
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		d, err := strconv.Atoi(part)
		if err == nil {
			days = append(days, d)
		}
	}
	return days
}

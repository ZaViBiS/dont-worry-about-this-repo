package bot

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ZaViBiS/dont-worry-about-this-repo/internal/db"

	charts "github.com/vicanso/go-charts/v2"
)

var monthMap = map[string]time.Month{
	"1": time.January, "01": time.January, "jan": time.January, "january": time.January, "січ": time.January, "січень": time.January, "січня": time.January,
	"2": time.February, "02": time.February, "feb": time.February, "february": time.February, "лют": time.February, "лютий": time.February, "лютого": time.February,
	"3": time.March, "03": time.March, "mar": time.March, "march": time.March, "бер": time.March, "березень": time.March, "березня": time.March,
	"4": time.April, "04": time.April, "apr": time.April, "april": time.April, "кві": time.April, "квітень": time.April, "квітня": time.April,
	"5": time.May, "05": time.May, "may": time.May, "тра": time.May, "травень": time.May, "травня": time.May,
	"6": time.June, "06": time.June, "jun": time.June, "june": time.June, "чер": time.June, "червень": time.June, "червня": time.June,
	"7": time.July, "07": time.July, "jul": time.July, "july": time.July, "лип": time.July, "липень": time.July, "липня": time.July,
	"8": time.August, "08": time.August, "aug": time.August, "august": time.August, "сер": time.August, "серпень": time.August, "серпня": time.August,
	"9": time.September, "09": time.September, "sep": time.September, "sept": time.September, "september": time.September, "вер": time.September, "вересень": time.September, "вересня": time.September,
	"10": time.October, "oct": time.October, "october": time.October, "жов": time.October, "жовтень": time.October, "жовтня": time.October,
	"11": time.November, "nov": time.November, "november": time.November, "лис": time.November, "листопад": time.November, "листопада": time.November,
	"12": time.December, "dec": time.December, "december": time.December, "гру": time.December, "грудень": time.December, "грудня": time.December,
}

var ukrMonthNames = map[time.Month]string{
	time.January:   "січень",
	time.February:  "лютий",
	time.March:     "березень",
	time.April:     "квітень",
	time.May:       "травень",
	time.June:      "червень",
	time.July:      "липень",
	time.August:    "серпень",
	time.September: "вересень",
	time.October:   "жовтень",
	time.November:  "листопад",
	time.December:  "грудень",
}

func parseMonth(s string) (time.Month, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	m, ok := monthMap[s]
	return m, ok
}

func parseMonthAndYearTokens(t1, t2 string) (time.Month, int, error) {
	m1, ok1 := parseMonth(t1)
	y2, err2 := strconv.Atoi(t2)
	if ok1 && err2 == nil && y2 >= 1000 && y2 <= 9999 {
		return m1, y2, nil
	}

	y1, err1 := strconv.Atoi(t1)
	m2, ok2 := parseMonth(t2)
	if ok2 && err1 == nil && y1 >= 1000 && y1 <= 9999 {
		return m2, y1, nil
	}

	return 0, 0, fmt.Errorf("cannot parse month and year from %s, %s", t1, t2)
}

func truncateToDay(t time.Time, loc *time.Location) time.Time {
	inLoc := t.In(loc)
	return time.Date(inLoc.Year(), inLoc.Month(), inLoc.Day(), 0, 0, 0, 0, loc)
}

// parseDateRange парсить аргументи команди stat та повертає діапазон дат і назву періоду.
func parseDateRange(args []string, now time.Time, loc *time.Location, records []db.Record) (time.Time, time.Time, string, error) {
	nowInLoc := now.In(loc)
	currYear := nowInLoc.Year()
	currMonth := nowInLoc.Month()

	if len(args) == 0 {
		startDate := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, loc)
		endDate := startDate.AddDate(0, 1, -1)
		return startDate, endDate, ukrMonthNames[currMonth], nil
	}

	var targetMonth time.Month
	var targetYear int

	if len(args) == 1 {
		arg := strings.TrimSpace(strings.ToLower(args[0]))
		var separator string
		if strings.Contains(arg, "-") {
			separator = "-"
		} else if strings.Contains(arg, ".") {
			separator = "."
		} else if strings.Contains(arg, "/") {
			separator = "/"
		}

		if separator != "" {
			parts := strings.Split(arg, separator)
			if len(parts) == 2 {
				m, y, err := parseMonthAndYearTokens(parts[0], parts[1])
				if err == nil {
					targetMonth = m
					targetYear = y
				}
			}
		}

		if targetMonth == 0 {
			m, ok := parseMonth(arg)
			if !ok {
				return time.Time{}, time.Time{}, "", fmt.Errorf("невідомий місяць '%s'. Приклад: stat, stat dec, stat 10, stat грудень", args[0])
			}
			targetMonth = m
			targetYear = currYear
			if targetMonth > currMonth {
				hasInCurrent := false
				for _, r := range records {
					t := r.Timestamp.In(loc)
					if t.Year() == currYear && t.Month() == targetMonth {
						hasInCurrent = true
						break
					}
				}
				if !hasInCurrent {
					targetYear = currYear - 1
				}
			}
		}
	} else if len(args) == 2 {
		m, y, err := parseMonthAndYearTokens(args[0], args[1])
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("некоректні аргументи: %s. Приклад: stat dec 2025", strings.Join(args, " "))
		}
		targetMonth = m
		targetYear = y
	} else {
		return time.Time{}, time.Time{}, "", fmt.Errorf("забагато аргументів: %s. Приклад: stat, stat dec, stat dec 2025", strings.Join(args, " "))
	}

	startDate := time.Date(targetYear, targetMonth, 1, 0, 0, 0, 0, loc)
	endDate := startDate.AddDate(0, 1, -1)

	var label string
	if targetYear == currYear {
		label = ukrMonthNames[targetMonth]
	} else {
		label = fmt.Sprintf("%s %d", ukrMonthNames[targetMonth], targetYear)
	}

	return startDate, endDate, label, nil
}

// generateChart генерує PNG лінійний графік кількості записів за днями.
func generateChart(records []db.Record, loc *time.Location, timeRange ...time.Time) ([]byte, int, error) {
	var startDate, endDate time.Time
	if len(timeRange) >= 2 {
		startDate = truncateToDay(timeRange[0], loc)
		endDate = truncateToDay(timeRange[1], loc)
	} else {
		if len(records) == 0 {
			return nil, 0, fmt.Errorf("no records to generate chart")
		}
		nowDate := truncateToDay(time.Now(), loc)
		startDate = time.Date(nowDate.Year(), nowDate.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1)
	}

	counts := make(map[string]int)
	for _, r := range records {
		dayKey := r.Timestamp.In(loc).Format("2006-01-02")
		counts[dayKey]++
	}

	var dates []string
	var values []float64
	var totalInPeriod int
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		c := counts[dayKey]
		dates = append(dates, d.Format("02.01"))
		values = append(values, float64(c))
		totalInPeriod += c
	}

	yAxis := map[string]any{
		"type": "value",
	}
	var maxVal float64
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		yAxis["min"] = 0
		yAxis["max"] = 1
	}

	opt := map[string]any{
		"legend": map[string]any{
			"show": false,
		},
		"title": map[string]any{
			"text":    "Статистика записів",
			"subtext": fmt.Sprintf("Всього записів: %d", totalInPeriod),
			"left":    "center",
		},
		"xAxis": map[string]any{
			"type": "category",
			"data": dates,
		},
		"yAxis": yAxis,
		"series": []map[string]any{
			{
				"name": "Записи",
				"type": "line",
				"data": values,
				"label": map[string]any{
					"show":     true,
					"position": "top",
				},
			},
		},
	}

	optBytes, err := json.Marshal(opt)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal chart options: %w", err)
	}

	pngBytes, err := charts.RenderEChartsToPNG(string(optBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("render echarts: %w", err)
	}

	return pngBytes, totalInPeriod, nil
}

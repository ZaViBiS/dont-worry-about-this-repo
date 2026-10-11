package bot

import (
	"encoding/json"
	"fmt"
	"time"

	"ZaViBiS/dont-worry-about-this-repo/internal/db"

	charts "github.com/vicanso/go-charts/v2"
)

// generateChart генерує PNG-графік кількості записів за днями.
func generateChart(records []db.Record, loc *time.Location) ([]byte, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to generate chart")
	}

	counts := make(map[string]int)
	for _, r := range records {
		dayKey := r.Timestamp.In(loc).Format("2006-01-02")
		counts[dayKey]++
	}

	firstDate := records[0].Timestamp.In(loc).Truncate(24 * time.Hour)
	lastDate := records[len(records)-1].Timestamp.In(loc).Truncate(24 * time.Hour)
	nowDate := time.Now().In(loc).Truncate(24 * time.Hour)

	endDate := nowDate
	if lastDate.After(endDate) {
		endDate = lastDate
	}

	daysDiff := int(endDate.Sub(firstDate).Hours() / 24)
	var startDate time.Time
	if daysDiff < 6 {
		startDate = endDate.AddDate(0, 0, -6)
	} else if daysDiff > 29 {
		startDate = endDate.AddDate(0, 0, -29)
	} else {
		startDate = firstDate
	}

	var dates []string
	var values []float64
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dayKey := d.Format("2006-01-02")
		dates = append(dates, d.Format("02.01"))
		values = append(values, float64(counts[dayKey]))
	}

	opt := map[string]any{
		"title": map[string]any{
			"text":    "Статистика записів",
			"subtext": fmt.Sprintf("Всього записів: %d", len(records)),
			"left":    "center",
		},
		"xAxis": map[string]any{
			"type": "category",
			"data": dates,
		},
		"yAxis": map[string]any{
			"type": "value",
		},
		"series": []map[string]any{
			{
				"name": "Записи",
				"type": "bar",
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
		return nil, fmt.Errorf("marshal chart options: %w", err)
	}

	pngBytes, err := charts.RenderEChartsToPNG(string(optBytes))
	if err != nil {
		return nil, fmt.Errorf("render echarts: %w", err)
	}

	return pngBytes, nil
}

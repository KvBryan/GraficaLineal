package chart_editor

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

// TransformChartSpecification adjusts series and metadata when switching chart types or updating table data
func TransformChartSpecification(spec *domain.ChartSpecification, targetType domain.ChartType) *domain.ChartSpecification {
	spec.ChartType = targetType

	switch targetType {
	case domain.ChartTypePie, domain.ChartTypeDoughnut:
		// Ensure series data has {name, value} structure
		if len(spec.Series) > 0 {
			var formattedData []interface{}
			for i, cat := range spec.Categories {
				var val float64 = 0
				if i < len(spec.Series[0].Data) {
					val = toFloat64(spec.Series[0].Data[i])
				}
				formattedData = append(formattedData, map[string]interface{}{
					"name":  cat,
					"value": val,
				})
			}
			spec.Series[0].Data = formattedData
			spec.Series[0].Type = "pie"
		}

	case domain.ChartTypeFunnel:
		if len(spec.Series) > 0 {
			var formattedData []interface{}
			for i, cat := range spec.Categories {
				var val float64 = 0
				if i < len(spec.Series[0].Data) {
					val = toFloat64(spec.Series[0].Data[i])
				}
				formattedData = append(formattedData, map[string]interface{}{
					"name":  cat,
					"value": val,
				})
			}
			spec.Series[0].Data = formattedData
			spec.Series[0].Type = "funnel"
		}

	case domain.ChartTypeRadar:
		// Ensure radar indicators exist
		if len(spec.RadarIndicators) == 0 && len(spec.Categories) > 0 {
			indicators := make([]domain.RadarIndicator, 0)
			for i, cat := range spec.Categories {
				var maxVal float64 = 100
				for _, s := range spec.Series {
					if i < len(s.Data) {
						v := toFloat64(s.Data[i])
						if v > maxVal {
							maxVal = v * 1.2
						}
					}
				}
				indicators = append(indicators, domain.RadarIndicator{
					Name: cat,
					Max:  math.Ceil(maxVal),
				})
			}
			spec.RadarIndicators = indicators
		}

	case domain.ChartTypeBoxplot:
		// If data is not in 5-number summary [min, Q1, median, Q3, max], calculate from table rows
		if len(spec.Series) > 0 && len(spec.TableUsed.Headers) >= 2 {
			calcBoxplotFromTable(spec)
		}

	case domain.ChartTypeHistogram:
		// Calculate bins and frequencies from the first numeric column of the table
		calcHistogramFromTable(spec)

	case domain.ChartTypeWaterfall:
		// In waterfall, categories are steps, data are deltas
		if len(spec.Series) > 0 {
			spec.Series[0].Type = "bar"
		}

	case domain.ChartTypeScatter:
		// Extract pairs (X, Y)
		if len(spec.TableUsed.Rows) > 0 {
			pairs := make([]interface{}, 0)
			for _, row := range spec.TableUsed.Rows {
				if len(row) >= 2 {
					xVal := toFloat64(row[0])
					yVal := toFloat64(row[1])
					pairs = append(pairs, []float64{xVal, yVal})
				}
			}
			spec.Series = []domain.SeriesData{
				{
					Name: "Observaciones (X vs Y)",
					Type: "scatter",
					Data: pairs,
				},
			}
		}

	case domain.ChartTypeHeatmap:
		if len(spec.HeatmapData) == 0 && len(spec.TableUsed.Rows) > 0 {
			calcHeatmapFromTable(spec)
		}
	}

	return spec
}

func calcBoxplotFromTable(spec *domain.ChartSpecification) {
	// Group numbers by category (Col 0) or by numeric columns
	colValues := make(map[string][]float64)

	// Check if each column is a group
	if len(spec.TableUsed.Headers) > 2 {
		for c := 1; c < len(spec.TableUsed.Headers); c++ {
			header := spec.TableUsed.Headers[c]
			for _, row := range spec.TableUsed.Rows {
				if c < len(row) {
					colValues[header] = append(colValues[header], toFloat64(row[c]))
				}
			}
		}
	} else {
		// Group by Col 0
		for _, row := range spec.TableUsed.Rows {
			if len(row) >= 2 {
				grp := fmt.Sprintf("%v", row[0])
				colValues[grp] = append(colValues[grp], toFloat64(row[1]))
			}
		}
	}

	categories := make([]string, 0)
	boxplotData := make([]interface{}, 0)

	for group, nums := range colValues {
		if len(nums) == 0 {
			continue
		}
		sort.Float64s(nums)
		minVal := nums[0]
		maxVal := nums[len(nums)-1]
		q1 := percentile(nums, 25)
		median := percentile(nums, 50)
		q3 := percentile(nums, 75)

		categories = append(categories, group)
		boxplotData = append(boxplotData, []float64{minVal, q1, median, q3, maxVal})
	}

	if len(categories) > 0 {
		spec.Categories = categories
		spec.Series = []domain.SeriesData{
			{
				Name: "Distribución (Min, Q1, Mediana, Q3, Max)",
				Type: "boxplot",
				Data: boxplotData,
			},
		}
	}
}

func calcHistogramFromTable(spec *domain.ChartSpecification) {
	var nums []float64
	for _, row := range spec.TableUsed.Rows {
		for c := 1; c < len(row); c++ {
			v := toFloat64(row[c])
			if v != 0 || row[c] == 0 || row[c] == 0.0 {
				nums = append(nums, v)
			}
		}
	}

	if len(nums) < 3 {
		return
	}

	sort.Float64s(nums)
	minVal := nums[0]
	maxVal := nums[len(nums)-1]
	if minVal == maxVal {
		maxVal += 1
	}

	binCount := int(math.Ceil(1 + 3.322*math.Log10(float64(len(nums)))))
	if binCount < 4 {
		binCount = 4
	}
	if binCount > 10 {
		binCount = 10
	}

	binWidth := (maxVal - minVal) / float64(binCount)
	bins := make([]int, binCount)
	categories := make([]string, binCount)

	for i := 0; i < binCount; i++ {
		start := minVal + float64(i)*binWidth
		end := start + binWidth
		categories[i] = fmt.Sprintf("%.1f - %.1f", start, end)
	}

	for _, n := range nums {
		idx := int((n - minVal) / binWidth)
		if idx >= binCount {
			idx = binCount - 1
		}
		if idx < 0 {
			idx = 0
		}
		bins[idx]++
	}

	freqData := make([]interface{}, binCount)
	for i, b := range bins {
		freqData[i] = b
	}

	spec.Categories = categories
	spec.XAxisName = "Intervalos de Frecuencia"
	spec.YAxisName = "Frecuencia Absoluta"
	spec.Series = []domain.SeriesData{
		{
			Name: "Frecuencia de Observaciones",
			Type: "bar",
			Data: freqData,
		},
	}
}

func calcHeatmapFromTable(spec *domain.ChartSpecification) {
	xLabels := spec.TableUsed.Headers[1:]
	yLabels := make([]string, 0)
	heatmapData := make([]domain.HeatmapPoint, 0)

	for y, row := range spec.TableUsed.Rows {
		if len(row) > 0 {
			yLabels = append(yLabels, fmt.Sprintf("%v", row[0]))
		}
		for x := 1; x < len(row); x++ {
			heatmapData = append(heatmapData, domain.HeatmapPoint{
				X:     x - 1,
				Y:     y,
				Value: toFloat64(row[x]),
			})
		}
	}

	spec.XAxisLabels = xLabels
	spec.YAxisLabels = yLabels
	spec.HeatmapData = heatmapData
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := (p / 100.0) * float64(len(sorted)-1)
	low := int(math.Floor(rank))
	high := int(math.Ceil(rank))
	weight := rank - float64(low)
	return sorted[low]*(1-weight) + sorted[high]*weight
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return 0
	case map[string]interface{}:
		if innerVal, ok := val["value"]; ok {
			return toFloat64(innerVal)
		}
		return 0
	default:
		return 0
	}
}

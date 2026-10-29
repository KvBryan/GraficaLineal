package chart_editor

import (
	"testing"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

func TestTransformChartSpecification(t *testing.T) {
	spec := &domain.ChartSpecification{
		ChartType:  domain.ChartTypeBar,
		Categories: []string{"A", "B", "C"},
		Series: []domain.SeriesData{
			{Name: "Serie 1", Data: []interface{}{10.0, 20.0, 30.0}},
		},
		TableUsed: domain.TableData{
			Headers: []string{"Cat", "Val"},
			Rows: [][]interface{}{
				{"A", 10.0},
				{"B", 20.0},
				{"C", 30.0},
			},
		},
	}

	// Transform to Pie
	pieSpec := TransformChartSpecification(spec, domain.ChartTypePie)
	if pieSpec.ChartType != domain.ChartTypePie {
		t.Fatalf("Expected chartType to be pie, got %s", pieSpec.ChartType)
	}

	// Transform to Radar
	radarSpec := TransformChartSpecification(spec, domain.ChartTypeRadar)
	if radarSpec.ChartType != domain.ChartTypeRadar {
		t.Fatalf("Expected chartType to be radar, got %s", radarSpec.ChartType)
	}
	if len(radarSpec.RadarIndicators) != 3 {
		t.Fatalf("Expected 3 radar indicators, got %d", len(radarSpec.RadarIndicators))
	}

	// Transform to Histogram
	histSpec := TransformChartSpecification(spec, domain.ChartTypeHistogram)
	if histSpec.ChartType != domain.ChartTypeHistogram {
		t.Fatalf("Expected chartType to be histogram, got %s", histSpec.ChartType)
	}
}

package domain

// ChartType represents one of the 13 supported scientific chart types
type ChartType string

const (
	// 1. Comparación entre categorías
	ChartTypeBar   ChartType = "bar"
	ChartTypeRadar ChartType = "radar"

	// 2. Evolución y cambio en el tiempo
	ChartTypeLine  ChartType = "line"
	ChartTypeArea  ChartType = "area"
	ChartTypeGantt ChartType = "gantt"

	// 3. Composición y partes de un todo
	ChartTypePie       ChartType = "pie"
	ChartTypeDoughnut  ChartType = "doughnut"
	ChartTypeFunnel    ChartType = "funnel"
	ChartTypeWaterfall ChartType = "waterfall"

	// 4. Distribución y relación
	ChartTypeScatter   ChartType = "scatter"
	ChartTypeHistogram ChartType = "histogram"
	ChartTypeBoxplot   ChartType = "boxplot"
	ChartTypeHeatmap   ChartType = "heatmap"
)

// SeriesData represents a single series plotted in a chart
type SeriesData struct {
	Name string        `json:"name"`
	Type string        `json:"type,omitempty"`
	Data []interface{} `json:"data"`
}

// RadarIndicator specifies indicator dimensions for radar charts
type RadarIndicator struct {
	Name string  `json:"name"`
	Max  float64 `json:"max,omitempty"`
}

// GanttItem defines a task bar in a Gantt diagram
type GanttItem struct {
	Task      string  `json:"task"`
	StartDate string  `json:"startDate"` // e.g. "2026-01-01" or step number
	EndDate   string  `json:"endDate"`   // e.g. "2026-03-01" or step number
	Duration  float64 `json:"durationDays"`
	Status    string  `json:"status,omitempty"`
}

// HeatmapPoint defines coordinates and value for 2D matrix
type HeatmapPoint struct {
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Value float64 `json:"value"`
}

// TableData stores the exact tabular data used for creating and auditing the chart
type TableData struct {
	Title          string          `json:"title"`
	Headers        []string        `json:"headers"`
	Rows           [][]interface{} `json:"rows"`
	SourceLocation string          `json:"sourceLocation"` // e.g. "Hoja1!A1:D10" or "Tabla 2 (pág. 3)"
}

// MethodologyData explains how the data was aggregated and prepared for research papers
type MethodologyData struct {
	Explanation        string            `json:"explanation"`        // Full explanation paragraph
	AggregationMethod  string            `json:"aggregationMethod"`  // e.g. "Promedios agrupados", "Conteo directo"
	VariablesMapping   map[string]string `json:"variablesMapping"`   // e.g. {"Eje X": "Mes", "Eje Y": "Tasa de Éxito"}
	AcademicCitationNote string          `json:"academicCitationNote"` // Citation ready for paper figure footer
}

// AIInformationBox specifies the explicit provenance and explanatory audit requested by researchers:
// 1. De qué información se hizo la gráfica
// 2. Qué datos usó exactamente
// 3. En dónde se encontraban los datos
// 4. Explicación detallada de la gráfica para el documento
type AIInformationBox struct {
	SourceDocument   string `json:"sourceDocument"`   // Archivo de origen
	SourceLocation   string `json:"sourceLocation"`   // En dónde (Hoja, rango de celdas o tabla de Word)
	DataUsedSummary  string `json:"dataUsedSummary"`  // Qué datos usó exactamente (filas, columnas, métricas)
	VariablesDetail  string `json:"variablesDetail"`  // Relación de variables (independientes, dependientes, factores)
	ChartExplanation string `json:"chartExplanation"` // Explicación de la gráfica (qué representa)
	KeyFindings      string `json:"keyFindings"`      // Hallazgos clave y tendencias cuantitativas
	PaperNarrative   string `json:"paperNarrative"`   // Narrativa redactada lista para insertar en el artículo
}

// ChartSpecification is the complete representation of an active chart
type ChartSpecification struct {
	ChartType        ChartType         `json:"chartType"`
	Title            string            `json:"title"`
	Subtitle         string            `json:"subtitle,omitempty"`
	XAxisName        string            `json:"xAxisName,omitempty"`
	YAxisName        string            `json:"yAxisName,omitempty"`
	Categories       []string          `json:"categories,omitempty"`
	Series           []SeriesData      `json:"series"`
	RadarIndicators  []RadarIndicator  `json:"radarIndicators,omitempty"`
	GanttItems       []GanttItem       `json:"ganttItems,omitempty"`
	HeatmapData      []HeatmapPoint    `json:"heatmapData,omitempty"`
	XAxisLabels      []string          `json:"xAxisLabels,omitempty"`
	YAxisLabels      []string          `json:"yAxisLabels,omitempty"`
	TableUsed        TableData         `json:"tableUsed"`
	Methodology      MethodologyData   `json:"methodology"`
	AIInfoBox        AIInformationBox  `json:"aiInfoBox"`
	ChartImageBase64 string            `json:"chartImageBase64,omitempty"` // Base64 PNG image for embedding in Excel
	Palette          string            `json:"palette,omitempty"`          // "apa", "navy", "pastel", "vibrant"
}

// DocumentTable represents a structured table detected in an Excel or Word file
type DocumentTable struct {
	Index       int             `json:"index"`
	Name        string          `json:"name"`
	Headers     []string        `json:"headers"`
	PreviewRows [][]interface{} `json:"previewRows"`
	TotalRows   int             `json:"totalRows"`
	TotalCols   int             `json:"totalCols"`
	SourceRef   string          `json:"sourceRef"` // e.g. "Hoja 'Resultados'" or "Word: Tabla 1"
}

// DocumentSection represents a paragraph or section heading in a Word document
type DocumentSection struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// DocumentInspectionResult is returned after uploading Word or Excel
type DocumentInspectionResult struct {
	FileType       string            `json:"fileType"` // "excel" or "word"
	FileName       string            `json:"fileName"`
	DocumentTitle  string            `json:"documentTitle,omitempty"`
	Summary        string            `json:"summary"`
	Tables         []DocumentTable   `json:"tables"`
	Sections       []DocumentSection `json:"sections,omitempty"`
	RecommendedType ChartType        `json:"recommendedType,omitempty"`
}

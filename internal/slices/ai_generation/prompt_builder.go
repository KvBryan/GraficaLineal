package ai_generation

import (
	"fmt"
	"strings"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

// GenerateRequest represents the client input to the AI slice
type GenerateRequest struct {
	DocumentType    string                `json:"documentType"`          // "excel" or "word"
	FileName        string                `json:"fileName"`
	SelectedTable   *domain.DocumentTable `json:"selectedTable,omitempty"`
	SelectedSection string                `json:"selectedSection,omitempty"`
	UserPrompt      string                `json:"userPrompt,omitempty"`
	RequestedType   string                `json:"requestedType,omitempty"` // one of the 13 types, or empty
}

func BuildSystemPrompt() string {
	return `Eres un analista de datos y bioestadístico experto en redacción científica para publicaciones académicas y tesis doctorales.
Tu misión es transformar tablas de datos o texto cuantitativo en especificaciones de visualización científica rigurosas, auditables y de alto impacto visual.

Debes responder SIEMPRE con un ÚNICO objeto JSON válido con la siguiente estructura exacta:
{
  "chartType": "bar | radar | line | area | gantt | pie | doughnut | funnel | waterfall | scatter | histogram | boxplot | heatmap",
  "title": "Título formal para artículo científico (ej. 'Figura 1: Comparativa de Rendimiento...')",
  "subtitle": "Subtítulo metodológico o unidades (ej. 'Valores expresados en porcentaje (n = 350)')",
  "xAxisName": "Nombre formal del Eje X con unidad si aplica",
  "yAxisName": "Nombre formal del Eje Y con unidad si aplica",
  "categories": ["Cat1", "Cat2", ...],
  "series": [
    {
      "name": "Nombre de la serie",
      "type": "bar | line | pie | scatter | etc",
      "data": [valores numéricos o arreglos según el tipo]
    }
  ],
  "radarIndicators": [{"name": "Dimensión A", "max": 100}], 
  "ganttItems": [{"task": "Fase 1", "startDate": "2026-01-01", "endDate": "2026-03-01", "durationDays": 60, "status": "Completado"}],
  "xAxisLabels": ["Col 1", "Col 2"],
  "yAxisLabels": ["Fila 1", "Fila 2"],
  "heatmapData": [{"x": 0, "y": 0, "value": 45.2}],
  "tableUsed": {
    "title": "Título descriptivo de la tabla de datos procesada",
    "headers": ["Columna1", "Columna2", ...],
    "rows": [["Fila1_Val1", 123], ["Fila2_Val1", 456]],
    "sourceLocation": "Ubicación exacta de la fuente (ej. 'Hoja Resultados!A1:D10' o 'Documento Word: Tabla 2')"
  },
  "methodology": {
    "explanation": "Explicación detallada de cómo se extrajeron y procesaron los datos para el artículo.",
    "aggregationMethod": "Método estadístico o de agregación (ej. 'Promedio ponderado', 'Mediana y RIQ', 'Frecuencia relativa (%)', 'Valores directos')",
    "variablesMapping": {
      "Variable X / Categoría": "Nombre de columna en origen",
      "Variable Y / Métrica": "Nombre de métrica en origen"
    },
    "academicCitationNote": "Nota al pie de figura para publicación (ej. 'Nota: Datos procesados a partir de la Tabla 2. p < 0.05.')"
  },
  "aiInfoBox": {
    "sourceDocument": "Nombre del archivo cargado",
    "sourceLocation": "Ubicación exacta de los datos (ej. 'Hoja Resultados, celdas A1:D15' o 'Documento Word: Tabla 2 en Sección Metodología')",
    "dataUsedSummary": "Descripción precisa de qué datos usó: qué filas, cuántas observaciones y qué columnas cuantitativas",
    "variablesDetail": "Detalle de variables: Variable X = '...', Series Y = '...'",
    "chartExplanation": "Explicación exhaustiva de la gráfica: qué representa, cómo interpretar los ejes y qué dinámica visual describe",
    "keyFindings": "Hallazgos principales y conclusiones analíticas reveladas por la visualización",
    "paperNarrative": "Párrafo académico formal redactado listo para insertar en los Resultados de la investigación (ej. 'Como se evidencia en la Figura 1...')"
  }
}

REGLAS ESPECÍFICAS SEGÚN TIPO DE GRÁFICA:
1. 'bar': 'categories' tiene las etiquetas del eje X. Cada elemento en 'series' tiene nombre y 'data' con arreglo de números.
2. 'radar': 'radarIndicators' tiene los indicadores con su 'max'. 'series[].data' tiene arreglos de números con la longitud de indicators.
3. 'line': 'categories' con puntos temporales/ordinales. 'series[].data' con números continuos.
4. 'area': Idem 'line' pero para acumulación o volumen.
5. 'gantt': Llenar 'ganttItems' con tareas, fechas y duraciones. 'categories' puede tener los nombres de las tareas.
6. 'pie' y 'doughnut': 'series[0].data' contiene objetos '[{"name": "...", "value": ...}, ...]'. 'categories' contiene los nombres.
7. 'funnel': 'series[0].data' contiene objetos '[{"name": "Etapa...", "value": ...}]' ordenados de mayor a menor.
8. 'waterfall': 'categories' contiene pasos (ej: ["Inicial", "+Ventas", "-Costos", "=Total"]). 'series[0].data' números (o deltas).
9. 'scatter': 'series[0].data' contiene pares '[ [x1, y1], [x2, y2], ... ]'.
10. 'histogram': 'categories' contiene los intervalos de clase (ej: ["0-10", "11-20", ...]) y 'series[0].data' contiene las frecuencias absolutas o relativas.
11. 'boxplot': 'categories' contiene los grupos analizados. 'series[0].data' contiene para cada grupo un arreglo de 5 números: [mínimo, Q1, mediana, Q3, máximo].
12. 'heatmap': 'xAxisLabels' columnas, 'yAxisLabels' filas, 'heatmapData' lista de {x, y, value}.

IMPORTANTE PARA AUDITORÍA Y METODOLOGÍA:
Debes ser 100% transparente: llena 'tableUsed' con las filas y columnas exactas que alimentan la gráfica.
En 'methodology.explanation', explica con precisión qué tablas o secciones se usaron y qué transformaciones matemáticas se aplicaron.`
}

func BuildUserPrompt(req *GenerateRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Nombre del archivo: %s (Tipo: %s)\n", req.FileName, req.DocumentType))

	if req.RequestedType != "" {
		sb.WriteString(fmt.Sprintf("Tipo de gráfica solicitado por el usuario: %s\n", req.RequestedType))
	} else {
		sb.WriteString("Elige el tipo de gráfica más adecuado para la naturaleza científica de estos datos.\n")
	}

	if req.UserPrompt != "" {
		sb.WriteString(fmt.Sprintf("Instrucción o hipótesis de investigación del usuario: %s\n", req.UserPrompt))
	}

	if req.SelectedTable != nil {
		sb.WriteString(fmt.Sprintf("\n--- TABLA SELECCIONADA (%s) ---\n", req.SelectedTable.SourceRef))
		sb.WriteString(fmt.Sprintf("Encabezados: %s\n", strings.Join(req.SelectedTable.Headers, " | ")))
		sb.WriteString("Filas de datos:\n")
		for i, row := range req.SelectedTable.PreviewRows {
			rowStr := make([]string, len(row))
			for j, cell := range row {
				rowStr[j] = fmt.Sprintf("%v", cell)
			}
			sb.WriteString(fmt.Sprintf("Fila %d: %s\n", i+1, strings.Join(rowStr, " | ")))
		}
	} else if req.SelectedSection != "" {
		sb.WriteString(fmt.Sprintf("\n--- SECCIÓN DE TEXTO SELECCIONADA ---\n%s\n", req.SelectedSection))
	}

	sb.WriteString("\nPor favor genera la especificación JSON completa con la gráfica, la tabla exacta usada y la metodología científica explicada.")
	return sb.String()
}

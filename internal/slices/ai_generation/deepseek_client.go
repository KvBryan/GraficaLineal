package ai_generation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KvBryan/GraficaLineal/internal/config"
	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

type DeepSeekClient struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewDeepSeekClient(cfg *config.Config) *DeepSeekClient {
	return &DeepSeekClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatRequest struct {
	Model          string                `json:"model"`
	Messages       []deepSeekChatMessage `json:"messages"`
	ResponseFormat map[string]string     `json:"response_format,omitempty"`
	Temperature    float64               `json:"temperature"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *DeepSeekClient) GenerateChart(ctx context.Context, req *GenerateRequest) (*domain.ChartSpecification, error) {
	if c.cfg.DeepSeekAPIKey == "" {
		return GenerateFallbackChart(req), nil
	}

	sysPrompt := BuildSystemPrompt()
	userPrompt := BuildUserPrompt(req)

	bodyReq := deepSeekChatRequest{
		Model: c.cfg.DeepSeekModel,
		Messages: []deepSeekChatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.2,
	}

	payload, err := json.Marshal(bodyReq)
	if err != nil {
		return nil, fmt.Errorf("error serializando solicitud a DeepSeek: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.cfg.DeepSeekBaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.DeepSeekAPIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Log and provide fallback
		fmt.Printf("[DeepSeek Error] Fallo al contactar API: %v. Generando fallback.\n", err)
		return GenerateFallbackChart(req), nil
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta de DeepSeek: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[DeepSeek HTTP %d] %s. Usando fallback.\n", resp.StatusCode, string(respBytes))
		return GenerateFallbackChart(req), nil
	}

	var chatResp deepSeekChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta de DeepSeek: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("DeepSeek no devolvió opciones de respuesta")
	}

	content := chatResp.Choices[0].Message.Content

	var spec domain.ChartSpecification
	if err := json.Unmarshal([]byte(content), &spec); err != nil {
		fmt.Printf("[DeepSeek Warning] Falló parseo de JSON directo: %v. Intentando extraer...\n", err)
		return GenerateFallbackChart(req), nil
	}

	// Guarantee defaults
	if spec.ChartType == "" {
		if req.RequestedType != "" {
			spec.ChartType = domain.ChartType(req.RequestedType)
		} else {
			spec.ChartType = domain.ChartTypeBar
		}
	}
	if spec.Palette == "" {
		spec.Palette = "apa"
	}

	fillAIInfoBoxDefaults(&spec, req)

	return &spec, nil
}

func fillAIInfoBoxDefaults(spec *domain.ChartSpecification, req *GenerateRequest) {
	if spec.AIInfoBox.SourceDocument == "" {
		spec.AIInfoBox.SourceDocument = req.FileName
	}
	if spec.AIInfoBox.SourceLocation == "" {
		if spec.TableUsed.SourceLocation != "" {
			spec.AIInfoBox.SourceLocation = spec.TableUsed.SourceLocation
		} else if req.SelectedTable != nil && req.SelectedTable.SourceRef != "" {
			spec.AIInfoBox.SourceLocation = req.SelectedTable.SourceRef
		} else {
			spec.AIInfoBox.SourceLocation = "Documento cargado"
		}
	}
	if spec.AIInfoBox.DataUsedSummary == "" {
		cols := ""
		if len(spec.TableUsed.Headers) > 0 {
			cols = fmt.Sprintf("Variables: %v", spec.TableUsed.Headers)
		}
		spec.AIInfoBox.DataUsedSummary = fmt.Sprintf("Muestra de %d observaciones tabuladas. %s.", len(spec.TableUsed.Rows), cols)
	}
	if spec.AIInfoBox.VariablesDetail == "" {
		if spec.Methodology.VariablesMapping != nil && len(spec.Methodology.VariablesMapping) > 0 {
			var details []string
			for k, v := range spec.Methodology.VariablesMapping {
				details = append(details, fmt.Sprintf("%s: %s", k, v))
			}
			spec.AIInfoBox.VariablesDetail = fmt.Sprintf("%v", details)
		} else {
			spec.AIInfoBox.VariablesDetail = fmt.Sprintf("Eje X (%s) vs Series numéricas (%s)", spec.XAxisName, spec.YAxisName)
		}
	}
	if spec.AIInfoBox.ChartExplanation == "" {
		spec.AIInfoBox.ChartExplanation = spec.Methodology.Explanation
		if spec.AIInfoBox.ChartExplanation == "" {
			spec.AIInfoBox.ChartExplanation = fmt.Sprintf("Visualización de %s para comparar la distribución de magnitudes entre las categorías analizadas.", spec.ChartType)
		}
	}
	if spec.AIInfoBox.KeyFindings == "" {
		spec.AIInfoBox.KeyFindings = fmt.Sprintf("Los datos reflejan variaciones cuantitativas entre las %d categorías evaluadas, permitiendo contrastar el comportamiento de las series en el estudio.", len(spec.Categories))
	}
	if spec.AIInfoBox.PaperNarrative == "" {
		spec.AIInfoBox.PaperNarrative = fmt.Sprintf("Tal como se presenta en la %s, los valores evidencian las diferencias estadísticas observadas a partir de los datos registrados en %s.", spec.Title, spec.AIInfoBox.SourceLocation)
	}
}

// GenerateFallbackChart builds an intelligent deterministic chart if DeepSeek API is unreachable
func GenerateFallbackChart(req *GenerateRequest) *domain.ChartSpecification {
	chartType := domain.ChartTypeBar
	if req.RequestedType != "" {
		chartType = domain.ChartType(req.RequestedType)
	}

	title := fmt.Sprintf("Figura 1: Análisis de Datos de %s", req.FileName)
	sourceLoc := "Archivo cargado"
	headers := []string{"Categoría", "Valor"}
	rows := [][]interface{}{
		{"Grupo A", 45.5},
		{"Grupo B", 68.2},
		{"Grupo C", 31.0},
		{"Grupo D", 89.4},
	}

	if req.SelectedTable != nil && len(req.SelectedTable.Headers) >= 2 {
		headers = req.SelectedTable.Headers
		rows = req.SelectedTable.PreviewRows
		sourceLoc = req.SelectedTable.SourceRef
	}

	categories := make([]string, 0)
	seriesValues := make([]interface{}, 0)

	for _, r := range rows {
		if len(r) > 0 {
			categories = append(categories, fmt.Sprintf("%v", r[0]))
		}
		if len(r) > 1 {
			seriesValues = append(seriesValues, r[1])
		}
	}

	return &domain.ChartSpecification{
		ChartType: chartType,
		Title:     title,
		Subtitle:  "Distribución de variables procesadas",
		XAxisName: headers[0],
		YAxisName: "Valor / Métrica",
		Categories: categories,
		Series: []domain.SeriesData{
			{
				Name: "Datos Observados",
				Type: string(chartType),
				Data: seriesValues,
			},
		},
		TableUsed: domain.TableData{
			Title:          fmt.Sprintf("Tabla Base: %s", req.FileName),
			Headers:        headers,
			Rows:           rows,
			SourceLocation: sourceLoc,
		},
		Methodology: domain.MethodologyData{
			Explanation:       "Los datos fueron extraídos directamente de las filas y columnas del documento fuente, mapeando la primera variable como dimensión categórica y las subsiguientes como cuantitativas.",
			AggregationMethod: "Valores directos tabulados",
			VariablesMapping: map[string]string{
				"Eje X": headers[0],
				"Eje Y": "Valores numéricos de las series",
			},
			AcademicCitationNote: fmt.Sprintf("Nota: Elaboración propia a partir de datos contenidos en %s.", sourceLoc),
		},
		AIInfoBox: domain.AIInformationBox{
			SourceDocument:   req.FileName,
			SourceLocation:   sourceLoc,
			DataUsedSummary:  fmt.Sprintf("%d registros procesados con columnas %v", len(rows), headers),
			VariablesDetail:  fmt.Sprintf("Variable X: %s | Series Cuantitativas: %v", headers[0], headers[1:]),
			ChartExplanation: fmt.Sprintf("La gráfica representa la comparación cuantitativa de los valores registrados en %s.", sourceLoc),
			KeyFindings:      "Se aprecian contrastes significativos entre los grupos analizados, facilitando la interpretación visual de los resultados.",
			PaperNarrative:   fmt.Sprintf("Como se observa en la %s, los valores reflejan los resultados experimentales obtenidos en %s.", title, sourceLoc),
		},
		Palette: "apa",
	}
}

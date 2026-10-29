package excel_export

import (
	"encoding/base64"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
	"time"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
	"github.com/xuri/excelize/v2"
)

// GenerateResearchExcelWorkbook creates a multi-sheet, beautifully styled Excel file with the chart, data and AI information box
func GenerateResearchExcelWorkbook(spec *domain.ChartSpecification, writer io.Writer) error {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Data & Chart
	sheetData := "Datos de la Gráfica"
	f.SetSheetName("Sheet1", sheetData)

	// Styles
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "1E3A8A"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return err
	}

	subtitleStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Size: 10, Color: "4B5563"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E40AF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	cellStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "E2E8F0", Style: 1},
			{Type: "bottom", Color: "E2E8F0", Style: 1},
			{Type: "left", Color: "E2E8F0", Style: 1},
			{Type: "right", Color: "E2E8F0", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// Write Title & Subtitle
	f.SetCellValue(sheetData, "A1", spec.Title)
	f.SetCellStyle(sheetData, "A1", "A1", titleStyle)

	sub := spec.Subtitle
	if sub == "" {
		sub = fmt.Sprintf("Generado para investigación científica el %s", time.Now().Format("02/01/2006"))
	}
	f.SetCellValue(sheetData, "A2", sub)
	f.SetCellStyle(sheetData, "A2", "A2", subtitleStyle)

	// Write Table Headers at Row 4
	startRow := 4
	headers := spec.TableUsed.Headers
	if len(headers) == 0 {
		headers = []string{"Categoría", "Valor"}
	}

	for colIdx, h := range headers {
		cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, startRow)
		f.SetCellValue(sheetData, cellRef, h)
		f.SetCellStyle(sheetData, cellRef, cellRef, headerStyle)
	}

	// Write Table Data Rows
	currentRow := startRow + 1
	for _, row := range spec.TableUsed.Rows {
		for colIdx, val := range row {
			if colIdx < len(headers) {
				cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, currentRow)
				f.SetCellValue(sheetData, cellRef, val)
				f.SetCellStyle(sheetData, cellRef, cellRef, cellStyle)
			}
		}
		currentRow++
	}

	// Footer note
	footerRow := currentRow + 1
	footerCell, _ := excelize.CoordinatesToCellName(1, footerRow)
	noteText := spec.Methodology.AcademicCitationNote
	if noteText == "" {
		noteText = fmt.Sprintf("Fuente original: %s", spec.TableUsed.SourceLocation)
	}
	f.SetCellValue(sheetData, footerCell, noteText)
	f.SetCellStyle(sheetData, footerCell, footerCell, subtitleStyle)

	// Placement column for chart & image (placed next to data table)
	startCol := len(headers) + 2
	chartColLetter, _ := excelize.ColumnNumberToName(startCol)
	chartCell := fmt.Sprintf("%s4", chartColLetter)

	// 1. Generate Native Interactive Excel Chart linked directly to table cells!
	if len(spec.TableUsed.Rows) > 0 && len(headers) >= 2 {
		lastRow := 4 + len(spec.TableUsed.Rows)
		catRange := fmt.Sprintf("'%s'!$A$5:$A$%d", sheetData, lastRow)

		var seriesList []excelize.ChartSeries
		for c := 2; c <= len(headers); c++ {
			colLetter, _ := excelize.ColumnNumberToName(c)
			seriesList = append(seriesList, excelize.ChartSeries{
				Name:       fmt.Sprintf("'%s'!$%s$4", sheetData, colLetter),
				Categories: catRange,
				Values:     fmt.Sprintf("'%s'!$%s$5:$%s$%d", sheetData, colLetter, colLetter, lastRow),
			})
		}

		var chartType excelize.ChartType
		switch spec.ChartType {
		case domain.ChartTypeBar:
			chartType = excelize.Col
		case domain.ChartTypeLine:
			chartType = excelize.Line
		case domain.ChartTypeArea:
			chartType = excelize.Area
		case domain.ChartTypePie:
			chartType = excelize.Pie
		case domain.ChartTypeDoughnut:
			chartType = excelize.Doughnut
		case domain.ChartTypeRadar:
			chartType = excelize.Radar
		case domain.ChartTypeScatter:
			chartType = excelize.Scatter
		default:
			chartType = excelize.Col
		}

		enable := true
		errChart := f.AddChart(sheetData, chartCell, &excelize.Chart{
			Type:   chartType,
			Series: seriesList,
			Title: excelize.ChartTitle{
				Paragraph: []excelize.RichTextRun{{Text: spec.Title}},
			},
			Legend: excelize.ChartLegend{
				Position: "bottom",
			},
			Format: excelize.GraphicOptions{
				ScaleX:      1.2,
				ScaleY:      1.2,
				PrintObject: &enable,
			},
		})
		if errChart != nil {
			fmt.Printf("[Excel Export Warning] Falló AddChart nativo: %v\n", errChart)
		}
	}

	// 2. Embed Rendered Graphic Picture (PNG from ECharts)
	if spec.ChartImageBase64 != "" {
		rawBase64 := spec.ChartImageBase64
		if idx := strings.Index(rawBase64, ","); idx != -1 {
			rawBase64 = rawBase64[idx+1:]
		}
		pngBytes, err := base64.StdEncoding.DecodeString(rawBase64)
		if err == nil && len(pngBytes) > 0 {
			// Place below native chart (row 22) or at chartCell if native chart could not be created
			picCell := fmt.Sprintf("%s22", chartColLetter)
			if len(headers) < 2 || len(spec.TableUsed.Rows) == 0 {
				picCell = chartCell
			}
			enable := true
			errPic := f.AddPictureFromBytes(sheetData, picCell, &excelize.Picture{
				Extension: ".png",
				File:      pngBytes,
				Format: &excelize.GraphicOptions{
					ScaleX:          0.65,
					ScaleY:          0.65,
					LockAspectRatio: true,
					AltText:         spec.Title,
					PrintObject:     &enable,
				},
			})
			if errPic != nil {
				fmt.Printf("[Excel Export Warning] Falló AddPictureFromBytes: %v\n", errPic)
			}
		} else if err != nil {
			fmt.Printf("[Excel Export Warning] Error decodificando imagen base64: %v\n", err)
		}
	}

	// Adjust column widths for Sheet 1
	for colIdx := range headers {
		colLetter, _ := excelize.ColumnNumberToName(colIdx + 1)
		f.SetColWidth(sheetData, colLetter, colLetter, 22)
	}

	// Sheet 2: Cuadro de Información de la IA y Auditoría Metodológica
	sheetMethod := "Cuadro de Información IA"
	f.NewSheet(sheetMethod)

	f.SetCellValue(sheetMethod, "A1", "📌 CUADRO DE INFORMACIÓN Y AUDITORÍA DE LA IA")
	f.SetCellStyle(sheetMethod, "A1", "A1", titleStyle)

	f.SetCellValue(sheetMethod, "A2", "Detalle exacto de la información de origen, datos usados, ubicación y explicación de la gráfica.")
	f.SetCellStyle(sheetMethod, "A2", "A2", subtitleStyle)

	sectionHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1E3A8A"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	metaKeyStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Color: "1E293B"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F1F5F9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	narrativeStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "0F172A"},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border: []excelize.Border{
			{Type: "top", Color: "CBD5E1", Style: 1},
			{Type: "bottom", Color: "CBD5E1", Style: 1},
			{Type: "left", Color: "CBD5E1", Style: 1},
			{Type: "right", Color: "CBD5E1", Style: 1},
		},
	})

	mRow := 4

	// Helper to add a section block
	addSection := func(title string, rows [][]string) {
		f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), title)
		f.SetCellValue(sheetMethod, fmt.Sprintf("B%d", mRow), "")
		f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("B%d", mRow), sectionHeaderStyle)
		mRow++

		for _, r := range rows {
			f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), r[0])
			f.SetCellValue(sheetMethod, fmt.Sprintf("B%d", mRow), r[1])
			f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("A%d", mRow), metaKeyStyle)
			f.SetCellStyle(sheetMethod, fmt.Sprintf("B%d", mRow), fmt.Sprintf("B%d", mRow), narrativeStyle)
			mRow++
		}
		mRow++
	}

	// 1. ¿De qué información se hizo la gráfica?
	addSection("1. ¿DE QUÉ INFORMACIÓN SE HIZO LA GRÁFICA? (FUENTE DE ORIGEN)", [][]string{
		{"Documento / Archivo Fuente", spec.AIInfoBox.SourceDocument},
		{"Título de la Figura / Gráfica", spec.Title},
		{"Subtítulo / Muestra", spec.Subtitle},
		{"Tipo de Gráfica Seleccionada", string(spec.ChartType)},
		{"Fecha y Hora de Generación", time.Now().Format("2006-01-02 15:04:05 MST")},
	})

	// 2. ¿En dónde se encontraban los datos?
	addSection("2. ¿EN DÓNDE SE ENCONTRABAN LOS DATOS? (UBICACIÓN Y TRAZABILIDAD)", [][]string{
		{"Ubicación Exacta en el Archivo", spec.AIInfoBox.SourceLocation},
		{"Rango de Celdas o Tabla Fuente", spec.TableUsed.SourceLocation},
		{"Encabezados de Origen Detectados", fmt.Sprintf("%v", spec.TableUsed.Headers)},
		{"Volumen de Datos", fmt.Sprintf("%d filas × %d columnas", len(spec.TableUsed.Rows), len(spec.TableUsed.Headers))},
	})

	// 3. ¿Qué datos usó la IA?
	addSection("3. ¿QUÉ DATOS USÓ LA IA? (VARIABLES Y TRANSFORMACIÓN)", [][]string{
		{"Resumen de Datos Utilizados", spec.AIInfoBox.DataUsedSummary},
		{"Detalle de Variables Mapeadas", spec.AIInfoBox.VariablesDetail},
		{"Método Estadístico / Agregación", spec.Methodology.AggregationMethod},
	})

	// 4. Explicación de la Gráfica
	addSection("4. EXPLICACIÓN DETALLADA DE LA GRÁFICA (INTERPRETACIÓN Y RESULTADOS)", [][]string{
		{"¿Qué representa la gráfica?", spec.AIInfoBox.ChartExplanation},
		{"Hallazgos Clave y Tendencias", spec.AIInfoBox.KeyFindings},
		{"Narrativa para Documento de Investigación", spec.AIInfoBox.PaperNarrative},
		{"Cita Académica / Nota al Pie Sugerida", spec.Methodology.AcademicCitationNote},
	})

	// Variable mappings subtable if any
	if len(spec.Methodology.VariablesMapping) > 0 {
		f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), "Mapeo Detallado de Ejes y Columnas")
		f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("B%d", mRow), sectionHeaderStyle)
		mRow++

		for k, v := range spec.Methodology.VariablesMapping {
			f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), k)
			f.SetCellValue(sheetMethod, fmt.Sprintf("B%d", mRow), v)
			f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("A%d", mRow), metaKeyStyle)
			f.SetCellStyle(sheetMethod, fmt.Sprintf("B%d", mRow), fmt.Sprintf("B%d", mRow), narrativeStyle)
			mRow++
		}
	}

	f.SetColWidth(sheetMethod, "A", "A", 36)
	f.SetColWidth(sheetMethod, "B", "B", 85)

	return f.Write(writer)
}

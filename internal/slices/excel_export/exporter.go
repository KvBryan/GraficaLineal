package excel_export

import (
	"fmt"
	"io"
	"time"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
	"github.com/xuri/excelize/v2"
)

// GenerateResearchExcelWorkbook creates a multi-sheet, beautifully styled Excel file with the chart data and methodology
func GenerateResearchExcelWorkbook(spec *domain.ChartSpecification, writer io.Writer) error {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Data
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

	// Adjust column widths
	for colIdx := range headers {
		colLetter, _ := excelize.ColumnNumberToName(colIdx + 1)
		f.SetColWidth(sheetData, colLetter, colLetter, 22)
	}

	// Sheet 2: Methodology & Auditing
	sheetMethod := "Metodología y Fuentes"
	f.NewSheet(sheetMethod)

	f.SetCellValue(sheetMethod, "A1", "Auditoría Metodológica de Datos y Fuentes")
	f.SetCellStyle(sheetMethod, "A1", "A1", titleStyle)

	f.SetCellValue(sheetMethod, "A2", "Detalle exacto de las tablas, cálculos y transformaciones aplicadas para el documento de investigación.")
	f.SetCellStyle(sheetMethod, "A2", "A2", subtitleStyle)

	metaHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"334155"}, Pattern: 1},
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

	f.SetCellValue(sheetMethod, "A4", "Parámetro Metodológico")
	f.SetCellValue(sheetMethod, "B4", "Descripción / Valor Registrado")
	f.SetCellStyle(sheetMethod, "A4", "A4", metaHeaderStyle)
	f.SetCellStyle(sheetMethod, "B4", "B4", metaHeaderStyle)

	metaRows := [][]string{
		{"Título de la Figura / Gráfica", spec.Title},
		{"Tipo de Gráfica Seleccionada", string(spec.ChartType)},
		{"Ubicación de Fuente Original", spec.TableUsed.SourceLocation},
		{"Método de Agregación Estadística", spec.Methodology.AggregationMethod},
		{"Explicación Metodológica Detallada", spec.Methodology.Explanation},
		{"Cita Académica Sugerida", spec.Methodology.AcademicCitationNote},
		{"Fecha y Hora de Generación", time.Now().Format("2006-01-02 15:04:05 MST")},
	}

	mRow := 5
	for _, m := range metaRows {
		f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), m[0])
		f.SetCellValue(sheetMethod, fmt.Sprintf("B%d", mRow), m[1])
		f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("A%d", mRow), metaKeyStyle)
		f.SetCellStyle(sheetMethod, fmt.Sprintf("B%d", mRow), fmt.Sprintf("B%d", mRow), cellStyle)
		mRow++
	}

	// Variables mapping subtable
	if len(spec.Methodology.VariablesMapping) > 0 {
		mRow++
		f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), "Mapeo de Variables Específicas")
		f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("B%d", mRow), metaHeaderStyle)
		mRow++

		for k, v := range spec.Methodology.VariablesMapping {
			f.SetCellValue(sheetMethod, fmt.Sprintf("A%d", mRow), k)
			f.SetCellValue(sheetMethod, fmt.Sprintf("B%d", mRow), v)
			f.SetCellStyle(sheetMethod, fmt.Sprintf("A%d", mRow), fmt.Sprintf("A%d", mRow), metaKeyStyle)
			f.SetCellStyle(sheetMethod, fmt.Sprintf("B%d", mRow), fmt.Sprintf("B%d", mRow), cellStyle)
			mRow++
		}
	}

	f.SetColWidth(sheetMethod, "A", "A", 35)
	f.SetColWidth(sheetMethod, "B", "B", 75)

	return f.Write(writer)
}

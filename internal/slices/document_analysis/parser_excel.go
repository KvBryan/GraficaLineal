package document_analysis

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
	"github.com/xuri/excelize/v2"
)

// ParseExcelFile parses an uploaded Excel file stream and returns inspection details
func ParseExcelFile(reader io.Reader, filename string) (*domain.DocumentInspectionResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo Excel: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("el archivo Excel no contiene hojas de cálculo")
	}

	result := &domain.DocumentInspectionResult{
		FileType: "excel",
		FileName: filename,
		Tables:   make([]domain.DocumentTable, 0),
	}

	totalRowsCount := 0
	for idx, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) == 0 {
			continue
		}

		// Find first non-empty row as header
		headerRowIdx := -1
		for rIdx, row := range rows {
			hasData := false
			for _, cell := range row {
				if strings.TrimSpace(cell) != "" {
					hasData = true
					break
				}
			}
			if hasData {
				headerRowIdx = rIdx
				break
			}
		}

		if headerRowIdx == -1 {
			continue
		}

		rawHeaders := rows[headerRowIdx]
		headers := make([]string, len(rawHeaders))
		for i, h := range rawHeaders {
			trimmed := strings.TrimSpace(h)
			if trimmed == "" {
				trimmed = fmt.Sprintf("Columna_%d", i+1)
			}
			headers[i] = trimmed
		}

		// Parse data rows
		dataRows := make([][]interface{}, 0)
		for r := headerRowIdx + 1; r < len(rows); r++ {
			row := rows[r]
			// check if row is empty
			isEmpty := true
			convertedRow := make([]interface{}, len(headers))
			for c := 0; c < len(headers); c++ {
				var valStr string
				if c < len(row) {
					valStr = strings.TrimSpace(row[c])
				}
				if valStr != "" {
					isEmpty = false
				}

				// Try parsing as number
				if valStr == "" {
					convertedRow[c] = ""
				} else if num, err := strconv.ParseFloat(strings.ReplaceAll(valStr, ",", "."), 64); err == nil {
					convertedRow[c] = num
				} else {
					convertedRow[c] = valStr
				}
			}
			if !isEmpty {
				dataRows = append(dataRows, convertedRow)
			}
		}

		previewLimit := 10
		if len(dataRows) < previewLimit {
			previewLimit = len(dataRows)
		}

		previewRows := dataRows[:previewLimit]
		totalRowsCount += len(dataRows)

		result.Tables = append(result.Tables, domain.DocumentTable{
			Index:       idx + 1,
			Name:        sheetName,
			Headers:     headers,
			PreviewRows: previewRows,
			TotalRows:   len(dataRows),
			TotalCols:   len(headers),
			SourceRef:   fmt.Sprintf("Hoja '%s' (Filas %d a %d)", sheetName, headerRowIdx+1, headerRowIdx+1+len(dataRows)),
		})
	}

	result.Summary = fmt.Sprintf("Excel analizado: %d hoja(s) con datos tabulares válidos y %d filas en total.", len(result.Tables), totalRowsCount)

	// Recommend initial chart type based on table shapes
	if len(result.Tables) > 0 {
		firstTbl := result.Tables[0]
		if firstTbl.TotalCols == 2 && firstTbl.TotalRows <= 7 {
			result.RecommendedType = domain.ChartTypePie
		} else if firstTbl.TotalCols > 2 {
			result.RecommendedType = domain.ChartTypeBar
		} else {
			result.RecommendedType = domain.ChartTypeLine
		}
	} else {
		result.RecommendedType = domain.ChartTypeBar
	}

	return result, nil
}

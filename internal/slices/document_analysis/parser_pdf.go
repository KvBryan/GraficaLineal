package document_analysis

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
	"github.com/ledongthuc/pdf"
)

var whitespaceRegex = regexp.MustCompile(`\s{2,}|\t`)

// ParsePDFFile parses an uploaded PDF file, extracts pages, detects tabular rows and sections
func ParsePDFFile(fileBytes []byte, filename string) (*domain.DocumentInspectionResult, error) {
	reader, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return nil, fmt.Errorf("error al abrir el documento PDF: %w", err)
	}

	numPages := reader.NumPage()
	if numPages == 0 {
		return nil, fmt.Errorf("el archivo PDF no contiene páginas legibles")
	}

	tables := make([]domain.DocumentTable, 0)
	sections := make([]domain.DocumentSection, 0)
	tableIdx := 1

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil || strings.TrimSpace(pageText) == "" {
			continue
		}

		trimmedPage := strings.TrimSpace(pageText)

		// Create section for this page
		sections = append(sections, domain.DocumentSection{
			Index:   pageNum,
			Title:   fmt.Sprintf("Página %d del PDF", pageNum),
			Content: trimmedPage,
		})

		// Detect tabular lines within the page
		lines := strings.Split(trimmedPage, "\n")
		var candidateTableRows [][]string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				if len(candidateTableRows) >= 2 {
					tbl := buildTableFromCandidateRows(tableIdx, pageNum, candidateTableRows)
					if tbl != nil {
						tables = append(tables, *tbl)
						tableIdx++
					}
					candidateTableRows = nil
				}
				continue
			}

			// Split by 2+ spaces, tab, or pipe
			var cols []string
			if strings.Contains(line, "|") {
				raw := strings.Split(line, "|")
				for _, c := range raw {
					if tc := strings.TrimSpace(c); tc != "" {
						cols = append(cols, tc)
					}
				}
			} else {
				cols = whitespaceRegex.Split(line, -1)
			}

			if len(cols) >= 2 {
				// Check if at least one column has a number
				hasNumber := false
				for _, c := range cols {
					cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(c, "%", ""), ",", "."))
					if _, err := strconv.ParseFloat(cleaned, 64); err == nil {
						hasNumber = true
						break
					}
				}
				if hasNumber || len(candidateTableRows) == 0 {
					candidateTableRows = append(candidateTableRows, cols)
				}
			} else {
				if len(candidateTableRows) >= 2 {
					tbl := buildTableFromCandidateRows(tableIdx, pageNum, candidateTableRows)
					if tbl != nil {
						tables = append(tables, *tbl)
						tableIdx++
					}
					candidateTableRows = nil
				}
			}
		}

		if len(candidateTableRows) >= 2 {
			tbl := buildTableFromCandidateRows(tableIdx, pageNum, candidateTableRows)
			if tbl != nil {
				tables = append(tables, *tbl)
				tableIdx++
			}
		}
	}

	result := &domain.DocumentInspectionResult{
		FileType: "pdf",
		FileName: filename,
		Tables:   tables,
		Sections: sections,
	}

	if len(tables) > 0 {
		result.Summary = fmt.Sprintf("Documento PDF analizado: %d páginas procesadas. Se identificaron %d tabla(s) cuantitativas y %d secciones de texto.", numPages, len(tables), len(sections))
		result.RecommendedType = domain.ChartTypeBar
	} else if len(sections) > 0 {
		result.Summary = fmt.Sprintf("Documento PDF analizado: %d páginas de texto cuantitativo extraídas. Selecciona la página o describe los datos que deseas graficar.", numPages)
		result.RecommendedType = domain.ChartTypeBar
	} else {
		result.Summary = "Documento PDF analizado, pero no contiene texto seleccionable (puede ser un PDF escaneado sin OCR)."
	}

	return result, nil
}

func buildTableFromCandidateRows(tableIdx, pageNum int, rawRows [][]string) *domain.DocumentTable {
	if len(rawRows) < 2 {
		return nil
	}

	headerRow := rawRows[0]
	maxCols := len(headerRow)
	for _, r := range rawRows {
		if len(r) > maxCols {
			maxCols = len(r)
		}
	}

	headers := make([]string, maxCols)
	for i := 0; i < maxCols; i++ {
		if i < len(headerRow) && strings.TrimSpace(headerRow[i]) != "" {
			headers[i] = strings.TrimSpace(headerRow[i])
		} else {
			headers[i] = fmt.Sprintf("Variable_%d", i+1)
		}
	}

	dataRows := make([][]interface{}, 0)
	for r := 1; r < len(rawRows); r++ {
		row := rawRows[r]
		convertedRow := make([]interface{}, maxCols)
		hasAnyValue := false
		for c := 0; c < maxCols; c++ {
			var valStr string
			if c < len(row) {
				valStr = strings.TrimSpace(row[c])
			}
			if valStr == "" {
				convertedRow[c] = ""
			} else {
				hasAnyValue = true
				cleaned := strings.ReplaceAll(strings.ReplaceAll(valStr, "%", ""), ",", ".")
				if num, err := strconv.ParseFloat(cleaned, 64); err == nil {
					convertedRow[c] = num
				} else {
					convertedRow[c] = valStr
				}
			}
		}
		if hasAnyValue {
			dataRows = append(dataRows, convertedRow)
		}
	}

	if len(dataRows) == 0 {
		return nil
	}

	previewLimit := 10
	if len(dataRows) < previewLimit {
		previewLimit = len(dataRows)
	}

	return &domain.DocumentTable{
		Index:       tableIdx,
		Name:        fmt.Sprintf("Tabla PDF (Pág. %d, #%d)", pageNum, tableIdx),
		Headers:     headers,
		PreviewRows: dataRows[:previewLimit],
		TotalRows:   len(dataRows),
		TotalCols:   len(headers),
		SourceRef:   fmt.Sprintf("PDF: Página %d (Tabla %d con %d filas)", pageNum, tableIdx, len(dataRows)),
	}
}

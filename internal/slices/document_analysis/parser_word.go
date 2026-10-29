package document_analysis

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

// ParseWordFile parses an uploaded Word (.docx) file and extracts sections and structured tables
func ParseWordFile(fileBytes []byte, filename string) (*domain.DocumentInspectionResult, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return nil, fmt.Errorf("el archivo no parece ser un documento Word (.docx) válido: %w", err)
	}

	var documentXMLFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			documentXMLFile = f
			break
		}
	}

	if documentXMLFile == nil {
		return nil, fmt.Errorf("no se encontró word/document.xml dentro del archivo .docx")
	}

	rc, err := documentXMLFile.Open()
	if err != nil {
		return nil, fmt.Errorf("error al leer contenido del documento Word: %w", err)
	}
	defer rc.Close()

	xmlData, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("error leyendo bytes xml: %w", err)
	}

	tables, sections, err := extractTablesAndSectionsFromXML(xmlData)
	if err != nil {
		return nil, err
	}

	result := &domain.DocumentInspectionResult{
		FileType: "word",
		FileName: filename,
		Tables:   tables,
		Sections: sections,
	}

	if len(tables) > 0 {
		result.Summary = fmt.Sprintf("Documento Word analizado con éxito: se encontraron %d tabla(s) estructuradas y %d secciones de texto. Selecciona qué tabla o sección deseas graficar.", len(tables), len(sections))
		result.RecommendedType = domain.ChartTypeBar
	} else if len(sections) > 0 {
		result.Summary = fmt.Sprintf("Documento Word analizado: no se encontraron tablas directas, pero se detectaron %d secciones de texto con posibles datos cuantitativos.", len(sections))
		result.RecommendedType = domain.ChartTypeBar
	} else {
		result.Summary = "Documento Word analizado, pero no contiene texto ni tablas estructuradas detectables."
	}

	return result, nil
}

func extractTablesAndSectionsFromXML(xmlData []byte) ([]domain.DocumentTable, []domain.DocumentSection, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))

	tables := make([]domain.DocumentTable, 0)
	sections := make([]domain.DocumentSection, 0)

	var (
		inTable     bool
		inRow       bool
		inCell      bool
		inText      bool

		currentTableRows [][]string
		currentRowCells  []string
		currentCellText  strings.Builder

		currentSectionText strings.Builder
		lastHeading        string
		sectionIdx         = 1
		tableIdx           = 1
	)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error parseando XML de Word: %w", err)
		}

		switch elem := token.(type) {
		case xml.StartElement:
			local := elem.Name.Local
			switch local {
			case "tbl":
				inTable = true
				currentTableRows = make([][]string, 0)
			case "tr":
				if inTable {
					inRow = true
					currentRowCells = make([]string, 0)
				}
			case "tc":
				if inRow {
					inCell = true
					currentCellText.Reset()
				}
			case "p":
				// start of paragraph
			case "t":
				inText = true
			}

		case xml.EndElement:
			local := elem.Name.Local
			switch local {
			case "t":
				inText = false
			case "p":
				if !inTable {
					// Add newline between paragraphs in text section
					currentSectionText.WriteString("\n")
				}
			case "tc":
				if inRow {
					inCell = false
					currentRowCells = append(currentRowCells, strings.TrimSpace(currentCellText.String()))
					currentCellText.Reset()
				}
			case "tr":
				if inTable {
					inRow = false
					if len(currentRowCells) > 0 {
						currentTableRows = append(currentTableRows, currentRowCells)
					}
				}
			case "tbl":
				inTable = false
				if len(currentTableRows) > 1 {
					// Has at least a header row and 1 data row
					rawHeaders := currentTableRows[0]
					headers := make([]string, len(rawHeaders))
					for i, h := range rawHeaders {
						trimmed := strings.TrimSpace(h)
						if trimmed == "" {
							trimmed = fmt.Sprintf("Columna_%d", i+1)
						}
						headers[i] = trimmed
					}

					dataRows := make([][]interface{}, 0)
					for r := 1; r < len(currentTableRows); r++ {
						row := currentTableRows[r]
						convertedRow := make([]interface{}, len(headers))
						for c := 0; c < len(headers); c++ {
							var val string
							if c < len(row) {
								val = strings.TrimSpace(row[c])
							}
							if val == "" {
								convertedRow[c] = ""
							} else if num, err := strconv.ParseFloat(strings.ReplaceAll(val, ",", "."), 64); err == nil {
								convertedRow[c] = num
							} else {
								convertedRow[c] = val
							}
						}
						dataRows = append(dataRows, convertedRow)
					}

					previewLimit := 10
					if len(dataRows) < previewLimit {
						previewLimit = len(dataRows)
					}

					tables = append(tables, domain.DocumentTable{
						Index:       tableIdx,
						Name:        fmt.Sprintf("Tabla %d", tableIdx),
						Headers:     headers,
						PreviewRows: dataRows[:previewLimit],
						TotalRows:   len(dataRows),
						TotalCols:   len(headers),
						SourceRef:   fmt.Sprintf("Documento Word: Tabla %d (%d filas × %d cols)", tableIdx, len(dataRows), len(headers)),
					})
					tableIdx++
				}
			}

		case xml.CharData:
			text := string(elem)
			if inText {
				if inCell {
					currentCellText.WriteString(text)
				} else if !inTable {
					currentSectionText.WriteString(text)
				}
			}
		}

		// Group non-table paragraphs into sections periodically or by length
		if !inTable && currentSectionText.Len() > 800 {
			content := strings.TrimSpace(currentSectionText.String())
			if len(content) > 50 {
				title := lastHeading
				if title == "" {
					title = fmt.Sprintf("Sección %d del Documento", sectionIdx)
				}
				sections = append(sections, domain.DocumentSection{
					Index:   sectionIdx,
					Title:   title,
					Content: content,
				})
				sectionIdx++
			}
			currentSectionText.Reset()
		}
	}

	// Any trailing section text
	if currentSectionText.Len() > 50 {
		sections = append(sections, domain.DocumentSection{
			Index:   sectionIdx,
			Title:   fmt.Sprintf("Sección %d del Documento", sectionIdx),
			Content: strings.TrimSpace(currentSectionText.String()),
		})
	}

	return tables, sections, nil
}

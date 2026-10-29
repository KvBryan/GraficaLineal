package excel_export

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/chart/export-excel", h.ExportExcel)
}

func (h *Handler) ExportExcel(w http.ResponseWriter, r *http.Request) {
	var spec domain.ChartSpecification
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "JSON inválido en solicitud de exportación: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Clean filename
	cleanTitle := sanitizeFilename(spec.Title)
	if cleanTitle == "" {
		cleanTitle = "grafica_investigacion"
	}
	filename := fmt.Sprintf("%s_%s.xlsx", cleanTitle, time.Now().Format("20060102_150405"))

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := GenerateResearchExcelWorkbook(&spec, w); err != nil {
		http.Error(w, "Error generando libro Excel: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_\-]+`)

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = nonAlphanumericRegex.ReplaceAllString(s, "")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

package chart_editor

import (
	"encoding/json"
	"net/http"

	"github.com/KvBryan/GraficaLineal/internal/core/domain"
)

type TransformRequest struct {
	ChartSpec  domain.ChartSpecification `json:"chartSpec"`
	TargetType domain.ChartType          `json:"targetType"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/chart/transform", h.TransformChart)
}

func (h *Handler) TransformChart(w http.ResponseWriter, r *http.Request) {
	var req TransformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido en solicitud de transformación: "+err.Error(), http.StatusBadRequest)
		return
	}

	transformed := TransformChartSpecification(&req.ChartSpec, req.TargetType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transformed)
}

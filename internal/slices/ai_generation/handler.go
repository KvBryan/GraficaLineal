package ai_generation

import (
	"encoding/json"
	"net/http"

	"github.com/KvBryan/GraficaLineal/internal/config"
)

type Handler struct {
	client *DeepSeekClient
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		client: NewDeepSeekClient(cfg),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/chart/generate", h.GenerateChart)
}

func (h *Handler) GenerateChart(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido en el cuerpo de la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}

	chartSpec, err := h.client.GenerateChart(r.Context(), &req)
	if err != nil {
		http.Error(w, "Error generando gráfica con IA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chartSpec)
}

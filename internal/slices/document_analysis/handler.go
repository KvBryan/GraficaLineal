package document_analysis

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/document/inspect", h.InspectDocument)
}

func (h *Handler) InspectDocument(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 25MB
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "Archivo demasiado grande (máximo 25MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No se proporcionó ningún archivo en el campo 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))

	switch ext {
	case ".xlsx", ".xls":
		result, err := ParseExcelFile(file, header.Filename)
		if err != nil {
			http.Error(w, "Error analizando archivo Excel: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case ".docx":
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error al leer documento Word: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result, err := ParseWordFile(fileBytes, header.Filename)
		if err != nil {
			http.Error(w, "Error analizando documento Word: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	case ".pdf":
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error al leer documento PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result, err := ParsePDFFile(fileBytes, header.Filename)
		if err != nil {
			http.Error(w, "Error analizando documento PDF: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Formato no soportado. Por favor sube un archivo Excel (.xlsx, .xls), Word (.docx) o PDF (.pdf)", http.StatusBadRequest)
	}
}

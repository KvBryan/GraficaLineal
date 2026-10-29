package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/KvBryan/GraficaLineal/internal/config"
	"github.com/KvBryan/GraficaLineal/internal/slices/ai_generation"
	"github.com/KvBryan/GraficaLineal/internal/slices/chart_editor"
	"github.com/KvBryan/GraficaLineal/internal/slices/document_analysis"
	"github.com/KvBryan/GraficaLineal/internal/slices/excel_export"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// Register Vertical Slices
	docHandler := document_analysis.NewHandler()
	docHandler.RegisterRoutes(mux)

	aiHandler := ai_generation.NewHandler(cfg)
	aiHandler.RegisterRoutes(mux)

	chartHandler := chart_editor.NewHandler()
	chartHandler.RegisterRoutes(mux)

	exportHandler := excel_export.NewHandler()
	exportHandler.RegisterRoutes(mux)

	// Static Files & Web UI
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}

	staticDir := filepath.Join(workDir, "web", "static")
	staticFS := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticFS))

	// Serve Index HTML for exact root
	indexPath := filepath.Join(workDir, "web", "index.html")
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexPath)
	})

	// Healthcheck endpoint for Docker / cloud deployments
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown channel
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("=====================================================")
		log.Printf("🚀 Servidor GraficaLineal iniciado exitosamente")
		log.Printf("🌐 URL Local: http://localhost:%s", cfg.Port)
		log.Printf("🧠 Modelo IA: DeepSeek (%s)", cfg.DeepSeekModel)
		log.Printf("📁 Arquitectura: Vertical Slice Architecture en Go")
		log.Printf("=====================================================")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en el servidor HTTP: %v", err)
		}
	}()

	<-stopChan
	log.Println("Deteniendo el servidor ordenadamente...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error forzando el cierre del servidor: %v", err)
	}

	fmt.Println("Servidor cerrado correctamente.")
}

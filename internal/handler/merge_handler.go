package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"pdfmerge/internal/model"
	"pdfmerge/internal/repository"
	"time"

	"go.uber.org/zap"
)

type MergeHandler struct {
	Repo   repository.PDFRepository
	Logger *zap.SugaredLogger
}

func NewMergeHandler(repo repository.PDFRepository, logger *zap.SugaredLogger) *MergeHandler {
	return &MergeHandler{Repo: repo, Logger: logger}
}

func (h *MergeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		h.Logger.Warnf("Rejected non-POST method: %s", r.Method)
		return
	}

	var req model.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		h.Logger.Warnf("Invalid JSON body: %v", err)
		return
	}

	h.Logger.Infow("Starting merge", "name", req.Name, "url_count", len(req.URLs))

	pdfBytes, err := h.Repo.DownloadAndMerge(req.URLs)
	if err != nil {
		http.Error(w, "Merge failed: "+err.Error(), http.StatusInternalServerError)
		h.Logger.Errorw("Merge failed", "error", err)
		return
	}

	name := req.Name
	if name == "" {
		name = "merged_" + time.Now().Format("20060102150405")
	}
	if filepath.Ext(name) != ".pdf" {
		name += ".pdf"
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	_, err = w.Write(pdfBytes)
	if err != nil {
		h.Logger.Errorw("Failed to write response", "error", err)
	} else {
		h.Logger.Infow("PDF merged and sent", "filename", name, "size_bytes", len(pdfBytes))
	}
}

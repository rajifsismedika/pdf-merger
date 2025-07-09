package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"pdfmerge/internal/model"
	"pdfmerge/internal/repository"
	"time"

	"go.uber.org/zap"
)

type ReportHandler struct {
	Repo    repository.PDFRepository
	Logger  *zap.SugaredLogger
	BaseURL string
}

func NewReportHandler(repo repository.PDFRepository, logger *zap.SugaredLogger, baseURL string) *ReportHandler {
	return &ReportHandler{
		Repo:    repo,
		Logger:  logger,
		BaseURL: baseURL,
	}
}

// ServeHTTP handles GET /report/{id}
func (h *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		h.Logger.Warnf("Rejected non-GET method: %s", r.Method)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing report ID", http.StatusBadRequest)
		h.Logger.Warn("Missing ID in path")
		return
	}

	reportURL := fmt.Sprintf("%s/one-api/tools/merge_report/get/%s", h.BaseURL, id)
	resp, err := http.Get(reportURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch report metadata", http.StatusBadGateway)
		h.Logger.Errorw("Failed to fetch report data", "url", reportURL, "error", err)
		return
	}
	defer resp.Body.Close()
	// Read the body into a byte slice to log the raw response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		h.Logger.Errorw("Error reading response body", "url", reportURL, "error", err)
		return
	}

	// Log the raw response body
	h.Logger.Infow("Raw response body", "body", string(bodyBytes))

	// Decode the response body into the model.MergeRequest
	var reqWrap model.MergeWrapper
	if err := json.Unmarshal(bodyBytes, &reqWrap); err != nil {
		http.Error(w, "Invalid report response", http.StatusInternalServerError)
		h.Logger.Errorw("Invalid JSON from report API", "error", err)
		return
	}
	req := reqWrap.Data
	h.Logger.Infow("Processing report merge", "id", id, "name", req.Name, "url_count", len(req.URLs))

	fullURLs := make([]string, len(req.URLs))
	for i, u := range req.URLs {
		fullURLs[i] = h.BaseURL + ensureLeadingSlash(u)
	}

	pdfBytes, err := h.Repo.DownloadAndMerge(fullURLs)
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
	w.Header().Set("Content-Disposition", "inline; filename=\""+path.Base(name)+"\"")
	_, err = w.Write(pdfBytes)
	if err != nil {
		h.Logger.Errorw("Failed to write response", "error", err)
	} else {
		h.Logger.Infow("PDF merged and sent", "filename", name, "size_bytes", len(pdfBytes))
	}
}

func ensureLeadingSlash(u string) string {
	if len(u) > 0 && u[0] != '/' {
		return "/" + u
	}
	return u
}

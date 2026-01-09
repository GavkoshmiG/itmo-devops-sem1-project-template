package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

func PostPricesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		archiveType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
		if archiveType == "" {
			archiveType = "zip"
		}
		if archiveType != "zip" && archiveType != "tar" {
			http.Error(w, "unsupported archive type", http.StatusBadRequest)
			return
		}

		if err := r.ParseMultipartForm(64 << 20); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		csvReader, err := extractCSVReader(file, archiveType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := importCSV(db, csvReader); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		stats, err := loadStats(db)
		if err != nil {
			http.Error(w, "failed to load stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func GetPricesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csvData, err := exportCSV(db)
		if err != nil {
			http.Error(w, "failed to export data", http.StatusInternalServerError)
			return
		}

		zipData, err := buildZip(csvData)
		if err != nil {
			http.Error(w, "failed to build archive", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(zipData)
	}
}

package app

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

type csvRow struct {
	name       string
	category   string
	price      float64
	createDate time.Time
}

func importCSV(db *sql.DB, input io.Reader) (statsResponse, error) {
	reader := csv.NewReader(input)

	var rows []csvRow
	line := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return statsResponse{}, errors.New("invalid csv data")
		}
		line++
		if line == 1 && isHeader(record) {
			continue
		}
		if len(record) < 5 {
			return statsResponse{}, errors.New("invalid csv row")
		}

		name := strings.TrimSpace(record[1])
		category := strings.TrimSpace(record[2])

		price, err := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		if err != nil {
			return statsResponse{}, errors.New("invalid price")
		}

		createDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[4]))
		if err != nil {
			return statsResponse{}, errors.New("invalid date")
		}

		rows = append(rows, csvRow{
			name:       name,
			category:   category,
			price:      price,
			createDate: createDate,
		})
	}

	tx, err := db.Begin()
	if err != nil {
		return statsResponse{}, errors.New("failed to begin transaction")
	}

	stmt, err := tx.Prepare(`INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		tx.Rollback()
		return statsResponse{}, errors.New("failed to prepare insert")
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(row.name, row.category, row.price, row.createDate); err != nil {
			tx.Rollback()
			return statsResponse{}, errors.New("failed to insert row")
		}
	}

	var stats statsResponse
	stats.TotalItems = len(rows)
	if err := tx.QueryRow(`
		SELECT
			COUNT(DISTINCT category) AS total_categories,
			COALESCE(SUM(price), 0) AS total_price
		FROM prices
	`).Scan(&stats.TotalCategories, &stats.TotalPrice); err != nil {
		tx.Rollback()
		return statsResponse{}, errors.New("failed to load stats")
	}

	if err := tx.Commit(); err != nil {
		return statsResponse{}, errors.New("failed to commit data")
	}
	return stats, nil
}

func exportCSV(db *sql.DB) ([]byte, error) {
	rows, err := db.Query(`SELECT id, name, category, price, create_date FROM prices ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
		return nil, err
	}

	type csvOutRow struct {
		id         int64
		name       string
		category   string
		price      float64
		createDate time.Time
	}
	var outRows []csvOutRow

	for rows.Next() {
		var row csvOutRow
		if err := rows.Scan(&row.id, &row.name, &row.category, &row.price, &row.createDate); err != nil {
			return nil, err
		}
		outRows = append(outRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, row := range outRows {
		record := []string{
			strconv.FormatInt(row.id, 10),
			row.name,
			row.category,
			strconv.FormatFloat(row.price, 'f', -1, 64),
			row.createDate.Format("2006-01-02"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isHeader(record []string) bool {
	if len(record) < 5 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(record[0]), "id") &&
		strings.EqualFold(strings.TrimSpace(record[1]), "name") &&
		strings.EqualFold(strings.TrimSpace(record[2]), "category") &&
		strings.EqualFold(strings.TrimSpace(record[3]), "price") &&
		strings.EqualFold(strings.TrimSpace(record[4]), "create_date")
}

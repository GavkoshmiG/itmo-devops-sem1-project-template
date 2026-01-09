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

func importCSV(db *sql.DB, input io.Reader) error {
	reader := csv.NewReader(input)

	tx, err := db.Begin()
	if err != nil {
		return errors.New("failed to begin transaction")
	}

	stmt, err := tx.Prepare(`INSERT INTO prices (id, name, category, price, create_date) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		tx.Rollback()
		return errors.New("failed to prepare insert")
	}
	defer stmt.Close()

	line := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return errors.New("invalid csv data")
		}
		line++
		if line == 1 && isHeader(record) {
			continue
		}
		if len(record) < 5 {
			tx.Rollback()
			return errors.New("invalid csv row")
		}

		id, err := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
		if err != nil {
			tx.Rollback()
			return errors.New("invalid id")
		}

		name := strings.TrimSpace(record[1])
		category := strings.TrimSpace(record[2])

		price, err := strconv.ParseFloat(strings.TrimSpace(record[3]), 64)
		if err != nil {
			tx.Rollback()
			return errors.New("invalid price")
		}

		createDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[4]))
		if err != nil {
			tx.Rollback()
			return errors.New("invalid date")
		}

		if _, err := stmt.Exec(id, name, category, price, createDate); err != nil {
			tx.Rollback()
			return errors.New("failed to insert row")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.New("failed to commit data")
	}
	return nil
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

	for rows.Next() {
		var (
			id         int64
			name       string
			category   string
			price      float64
			createDate time.Time
		)
		if err := rows.Scan(&id, &name, &category, &price, &createDate); err != nil {
			return nil, err
		}
		row := []string{
			strconv.FormatInt(id, 10),
			name,
			category,
			strconv.FormatFloat(price, 'f', -1, 64),
			createDate.Format("2006-01-02"),
		}
		if err := writer.Write(row); err != nil {
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

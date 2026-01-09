package app

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path"
	"strings"
)

func extractCSVReader(file io.Reader, archiveType string) (io.Reader, error) {
	switch archiveType {
	case "zip":
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, errors.New("failed to read zip data")
		}
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, errors.New("invalid zip archive")
		}
		if csvData, ok := readZipCSV(reader, "data.csv"); ok {
			return bytes.NewReader(csvData), nil
		}
		if csvData, ok := readZipAnyCSV(reader); ok {
			return bytes.NewReader(csvData), nil
		}
		return nil, errors.New("csv file not found in zip")
	case "tar":
		reader := tar.NewReader(file)
		var fallback []byte
		for {
			hdr, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, errors.New("invalid tar archive")
			}
			if hdr.Typeflag != tar.TypeReg {
				continue
			}
			base := path.Base(hdr.Name)
			if base == "data.csv" {
				csvData, err := io.ReadAll(reader)
				if err != nil {
					return nil, errors.New("failed to read data.csv")
				}
				return bytes.NewReader(csvData), nil
			}
			if fallback == nil && strings.HasSuffix(strings.ToLower(base), ".csv") {
				csvData, err := io.ReadAll(reader)
				if err != nil {
					return nil, errors.New("failed to read csv data")
				}
				fallback = csvData
			}
		}
		if fallback != nil {
			return bytes.NewReader(fallback), nil
		}
		return nil, errors.New("csv file not found in tar")
	default:
		return nil, errors.New("unsupported archive type")
	}
}

func readZipCSV(reader *zip.Reader, name string) ([]byte, bool) {
	for _, f := range reader.File {
		if path.Base(f.Name) != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, false
		}
		defer rc.Close()
		csvData, err := io.ReadAll(rc)
		if err != nil {
			return nil, false
		}
		return csvData, true
	}
	return nil, false
}

func readZipAnyCSV(reader *zip.Reader) ([]byte, bool) {
	for _, f := range reader.File {
		if !strings.HasSuffix(strings.ToLower(path.Base(f.Name)), ".csv") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, false
		}
		defer rc.Close()
		csvData, err := io.ReadAll(rc)
		if err != nil {
			return nil, false
		}
		return csvData, true
	}
	return nil, false
}

func buildZip(csvData []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	file, err := writer.Create("data.csv")
	if err != nil {
		writer.Close()
		return nil, err
	}
	if _, err := file.Write(csvData); err != nil {
		writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

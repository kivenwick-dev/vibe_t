package codec

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"banner-fingerprint/internal/model"
)

// Decode accepts a JSON array, a stream of JSON objects, or CSV with a fixed header.
func Decode(r io.Reader, format string) ([]model.BannerInput, error) {
	inputs := make([]model.BannerInput, 0)
	switch format {
	case "json":
		d := json.NewDecoder(r)
		if err := d.Decode(&inputs); err != nil {
			return nil, err
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("expected a single JSON array")
		}
	case "jsonl":
		d := json.NewDecoder(r)
		for {
			var item model.BannerInput
			err := d.Decode(&item)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("record %d: %w", len(inputs)+1, err)
			}
			inputs = append(inputs, item)
		}
	case "csv":
		reader := csv.NewReader(r)
		header, err := reader.Read()
		if err != nil {
			return nil, err
		}
		if len(header) != 3 || header[0] != "ip" || header[1] != "port" || header[2] != "banner" {
			return nil, fmt.Errorf("CSV header must be ip,port,banner")
		}
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			port, err := strconv.Atoi(row[1])
			if err != nil || port < 1 || port > 65535 {
				return nil, fmt.Errorf("record %d has invalid port %q", len(inputs)+1, row[1])
			}
			inputs = append(inputs, model.BannerInput{IP: row[0], Port: port, Banner: row[2]})
		}
	default:
		return nil, fmt.Errorf("unsupported input format %q", format)
	}
	return inputs, nil
}

func Encode(w io.Writer, results []model.FingerprintResult, format string, pretty bool) error {
	switch format {
	case "json", "jsonl":
		encoder := json.NewEncoder(w)
		if format == "json" {
			if pretty {
				encoder.SetIndent("", "  ")
			}
			return encoder.Encode(results)
		}
		for _, result := range results {
			if err := encoder.Encode(result); err != nil {
				return err
			}
		}
		return nil
	case "csv":
		writer := csv.NewWriter(w)
		if err := writer.Write([]string{"ip", "port", "protocol", "product", "version", "os_hint", "confidence"}); err != nil {
			return err
		}
		for _, r := range results {
			if err := writer.Write([]string{r.IP, strconv.Itoa(r.Port), r.Protocol, r.Product, r.Version, r.OSHint, strconv.FormatFloat(r.Confidence, 'f', -1, 64)}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

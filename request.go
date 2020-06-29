package httpkit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
)

// RequestDecoder decode request values
var RequestDecoder = schema.NewDecoder()

func init() {
	RequestDecoder.IgnoreUnknownKeys(true)
}

// ScanValues 从url.Values解析数据
func ScanValues(dst interface{}, values url.Values) error {
	if err := RequestDecoder.Decode(dst, values); err != nil {
		return fmt.Errorf("decode values, %w", err)
	}

	if _, err := govalidator.ValidateStruct(dst); err != nil {
		return fmt.Errorf("validate values, %w", err)
	}

	return nil
}

// ScanJSON decode json and validate
func ScanJSON(dst interface{}, r io.Reader) (int, error) {
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		return http.StatusNotAcceptable, fmt.Errorf("json decode, %w", err)
	}

	if _, err := govalidator.ValidateStruct(dst); err != nil {
		return http.StatusBadRequest, fmt.Errorf("validate json, %w", err)
	}

	return 0, nil
}

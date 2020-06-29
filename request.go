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

// MustScanValue deocde request values, panic when error
func MustScanValue(dst interface{}, values url.Values) {
	if err := ScanValues(dst, values); err != nil {
		panic(WrapError(err).
			WithStatus(http.StatusBadRequest).
			WithString(err.Error()))
	}
}

// ScanJSON decode json and validate
func ScanJSON(dst interface{}, r io.Reader) error {
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		return fmt.Errorf("json decode, %w", err)
	}

	if _, err := govalidator.ValidateStruct(dst); err != nil {
		return fmt.Errorf("validate values, %w", err)
	}
	return nil
}

// MustScanJSON json decode request body, panic when error
func MustScanJSON(dst interface{}, r io.Reader) {
	if err := ScanJSON(dst, r); err != nil {
		panic(WrapError(err).
			WithStatus(http.StatusBadRequest).
			WithString(err.Error()))
	}
}

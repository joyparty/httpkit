package httpkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Error http错误
type Error struct {
	Header http.Header
	Body   io.Reader
	code   int
	err    error
}

// NewError 创建http错误
func NewError(code int) *Error {
	return &Error{
		code: code,
	}
}

// WrapError 把其它错误转换为http错误
func WrapError(err error) *Error {
	return &Error{
		code: http.StatusInternalServerError,
		err:  err,
	}
}

func (e Error) Error() string {
	return http.StatusText(e.StatusCode())
}

// Unwrap returns cause error
func (e Error) Unwrap() error {
	return e.err
}

// StatusCode returns response status code
func (e Error) StatusCode() int {
	if code := e.code; code > 0 {
		return code
	}
	return http.StatusInternalServerError
}

// WithStatus set response status code
func (e *Error) WithStatus(code int) *Error {
	e.code = code
	return e
}

// WithHeader set response header
func (e *Error) WithHeader(key, value string) *Error {
	if e.Header == nil {
		e.Header = http.Header{}
	}
	e.Header.Set(key, value)
	return e
}

// WithBody set response body
func (e *Error) WithBody(r io.Reader) *Error {
	e.Body = r
	return e
}

// WithJSON set json content as response body
func (e *Error) WithJSON(v interface{}) error {
	e.WithHeader("Content-Type", "application/json")

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode error response, %w", err)
	}

	return e.WithBytes(data)
}

// WithString set response body content
func (e *Error) WithString(s string) *Error {
	e.WithHeader("Content-Type", "text/plain")
	e.Body = strings.NewReader(s)
	return e
}

// WithBytes set response body content
func (e *Error) WithBytes(data []byte) *Error {
	e.Body = bytes.NewReader(data)
	return e
}

// WriteError write error response
func WriteError(w http.ResponseWriter, httpError *Error) error {
	w.WriteHeader(httpError.StatusCode())

	if h := httpError.Header; h != nil {
		for key, values := range h {
			for _, value := range values {
				w.Header().Set(key, value)
				break
			}
		}
	}

	if r := httpError.Body; r != nil {
		if _, err := io.Copy(w, r); err != nil {
			return err
		}
	}

	return nil
}

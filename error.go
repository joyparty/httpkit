package httpkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
)

var (
	callerInitOnce sync.Once
	wrapFunc       string
)

const maximumCallerDepth int = 25

// Error http错误
type Error struct {
	header http.Header
	body   io.Reader
	code   int
	err    error
	caller *runtime.Frame
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
		code:   http.StatusInternalServerError,
		err:    err,
		caller: GetCaller(),
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

// Header returns response header
func (e Error) Header() http.Header {
	return e.header
}

// Body returns response body reader
func (e Error) Body() io.Reader {
	return e.body
}

// Caller returns WrapError caller
func (e Error) Caller() (*runtime.Frame, bool) {
	return e.caller, e.caller != nil
}

// WithStatus set response status code
func (e *Error) WithStatus(code int) *Error {
	e.code = code
	return e
}

// WithHeader set response header
func (e *Error) WithHeader(key, value string) *Error {
	if e.header == nil {
		e.header = http.Header{}
	}
	e.header.Set(key, value)
	return e
}

// WithBody set response body
func (e *Error) WithBody(r io.Reader) *Error {
	e.body = r
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
	e.body = strings.NewReader(s)
	return e
}

// WithBytes set response body content
func (e *Error) WithBytes(data []byte) *Error {
	e.body = bytes.NewReader(data)
	return e
}

// WriteError write error response
func WriteError(w http.ResponseWriter, httpError *Error) error {
	if h := httpError.Header(); h != nil {
		for key, values := range h {
			for _, value := range values {
				w.Header().Set(key, value)
				break
			}
		}
	}
	w.WriteHeader(httpError.StatusCode())

	if r := httpError.Body(); r != nil {
		if _, err := io.Copy(w, r); err != nil {
			return err
		}
	}

	return nil
}

// GetCaller 返回当前函数的调用者
func GetCaller() *runtime.Frame {
	pc := make([]uintptr, 2)
	n := runtime.Callers(3, pc)
	if n == 0 {
		panic(errors.New("unknown caller"))
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	for {
		frame, _ := frames.Next()
		return &frame
	}
}

package httpkit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// RequestLogger 请求日志记录，可替换
	RequestLogger = func(w http.ResponseWriter, r *http.Request, logger logrus.FieldLogger) logrus.FieldLogger {
		rl := r.URL

		fields := logrus.Fields{
			"remote": r.RemoteAddr,
			"method": r.Method,
			"path":   rl.Path,
		}

		if v := rl.RawQuery; v != "" {
			if q := rl.Query(); len(q) > 0 {
				fields["query"] = q
			}
		}

		if m := r.Method; m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch {
			r.ParseForm()
			if form := r.PostForm; len(form) > 0 {
				fields["form"] = form
			}
		}

		return logger.WithFields(fields)
	}

	responsePool = sync.Pool{
		New: func() interface{} {
			return &responseWrapper{}
		},
	}

	requestBodyPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

type responseWrapper struct {
	http.ResponseWriter
	status int
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw responseWrapper) StatusCode() int {
	if c := rw.status; c > 0 {
		return c
	}
	return http.StatusOK
}

// LogRequest http访问日志
func LogRequest(logger logrus.FieldLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := responsePool.Get().(*responseWrapper)
			ww.ResponseWriter = w
			ww.status = 0
			defer func() {
				ww.ResponseWriter = nil
				responsePool.Put(ww)
			}()

			start := time.Now()

			var rBody *bytes.Buffer
			if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				if n, _ := strconv.Atoi(r.Header.Get("Content-Length")); n > 0 && n <= 8192 {
					rBody = requestBodyPool.Get().(*bytes.Buffer)
					rBody.Reset()
					defer func() {
						requestBodyPool.Put(rBody)
					}()

					r.Body = io.NopCloser(io.TeeReader(r.Body, rBody))
				}
			}

			next.ServeHTTP(ww, r)

			status := ww.StatusCode()
			fl := RequestLogger(w, r, logger).
				WithFields(logrus.Fields{
					"duration": time.Since(start).Milliseconds(),
					"status":   status,
				})
			if rBody != nil && rBody.Len() > 0 {
				fl = fl.WithField("body", rBody.String())
			}

			if status >= 500 {
				fl.Error("http request")
			} else if status >= 400 {
				fl.Warn("http request")
			} else {
				fl.Info("http request")
			}
		})
	}
}

// Recoverer recover panic
func Recoverer(logger logrus.FieldLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					switch vv := v.(type) {
					case *Error:
						if err := errors.Unwrap(vv); err != nil {
							entry := logger.WithError(err)
							if f, ok := vv.Caller(); ok {
								entry = entry.WithFields(logrus.Fields{
									logrus.FieldKeyFile: fmt.Sprintf("%s:%d", f.File, f.Line),
									logrus.FieldKeyFunc: f.Function,
								})
							} else {
								entry = entry.WithFields(logrus.Fields{
									"method": r.Method,
									"uri":    r.URL.Path,
								})
							}

							if code := vv.StatusCode(); code >= http.StatusInternalServerError {
								entry.Error("recover http error")
							} else {
								entry.Debug("recover http error")
							}
						}

						if err := WriteError(w, vv); err != nil {
							logger.WithError(err).Error("send error response")
						}
						return
					case error:
						fields := logrus.Fields{
							"method": r.Method,
							"uri":    r.URL.Path,
						}

						if _, ok := vv.(runtime.Error); ok {
							fields["stack"] = string(debug.Stack())
						}

						logger.WithError(vv).
							WithFields(fields).
							Error("recover panic")
					default:
						logger.WithField("error", v).
							WithFields(logrus.Fields{
								"method": r.Method,
								"uri":    r.URL.Path,
							}).
							Error("recover panic")
					}

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

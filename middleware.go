package httpkit

import (
	"errors"
	"net/http"
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

			start := time.Now()
			defer func() {
				status := ww.StatusCode()

				fl := RequestLogger(w, r, logger).
					WithFields(logrus.Fields{
						"duration": time.Now().Sub(start).Milliseconds(),
						"status":   status,
					})

				if status >= 500 {
					fl.Error("http request")
				} else if status >= 400 {
					fl.Warn("http request")
				} else {
					fl.Info("http request")
				}

				ww.ResponseWriter = nil
				responsePool.Put(ww)
			}()

			next.ServeHTTP(ww, r)
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
							logger.WithError(vv).Error("recover panic")
						}

						WriteError(w, vv)
						return
					case error:
						logger.WithError(vv).Error("recover panic")
					default:
						logger.WithField("error", v).Error("recover panic")
					}

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

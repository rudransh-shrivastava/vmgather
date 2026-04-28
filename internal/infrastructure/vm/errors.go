package vm

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type ErrorKind string

const (
	ErrorKindUnknown       ErrorKind = "unknown"
	ErrorKindMissingRoute  ErrorKind = "missing_route"
	ErrorKindQueryTimeout  ErrorKind = "query_timeout"
	ErrorKindTooManySeries ErrorKind = "too_many_series"
	ErrorKindTransient     ErrorKind = "transient"
)

type APIError struct {
	StatusCode int
	Body       string
	Kind       ErrorKind
	Err        error
}

func (e *APIError) Error() string {
	body := strings.TrimSpace(e.Body)
	if body == "" {
		if e.Err != nil {
			return e.Err.Error()
		}
		return "API request failed"
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("unexpected status code %d: %s", e.StatusCode, body)
	}
	return body
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func ErrorKindOf(err error) ErrorKind {
	if err == nil {
		return ErrorKindUnknown
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Kind != ErrorKindUnknown {
		return apiErr.Kind
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout exceeded during query execution"),
		strings.Contains(msg, "timeout exceeded during the query"),
		strings.Contains(msg, "timeout exceeded while fetching data block"),
		strings.Contains(msg, "timeout exceeded before starting data export"),
		strings.Contains(msg, "context deadline exceeded"):
		return ErrorKindQueryTimeout
	case strings.Contains(msg, "the number of matching timeseries exceeds"),
		strings.Contains(msg, "cannot select more than -search.maxexportseries"):
		return ErrorKindTooManySeries
	case strings.Contains(msg, "missing route"),
		strings.Contains(msg, "unsupported path"),
		strings.Contains(msg, "unexpected status code 404"),
		strings.Contains(msg, " not found"):
		return ErrorKindMissingRoute
	}

	if isTransientError(err) {
		return ErrorKindTransient
	}

	return ErrorKindUnknown
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "unexpected eof")
}

func classifyBody(body string) ErrorKind {
	msg := strings.ToLower(strings.TrimSpace(body))
	switch {
	case strings.Contains(msg, "timeout exceeded during query execution"),
		strings.Contains(msg, "timeout exceeded during the query"),
		strings.Contains(msg, "timeout exceeded while fetching data block"),
		strings.Contains(msg, "timeout exceeded before starting data export"):
		return ErrorKindQueryTimeout
	case strings.Contains(msg, "the number of matching timeseries exceeds"),
		strings.Contains(msg, "cannot select more than -search.maxexportseries"):
		return ErrorKindTooManySeries
	case strings.Contains(msg, "missing route"),
		strings.Contains(msg, "unsupported path"),
		strings.Contains(msg, "not found"):
		return ErrorKindMissingRoute
	default:
		return ErrorKindUnknown
	}
}

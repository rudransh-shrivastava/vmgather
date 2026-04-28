package vm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VictoriaMetrics/vmgather/internal/domain"
)

func TestQueryDetectsMissingTenantPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("cannot parse accountID from \"api\""))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer srv.Close()

	client := NewClient(domain.VMConnection{URL: srv.URL})
	_, err := client.Query(context.Background(), "vm_app_version", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingTenantPath) {
		t.Fatalf("expected ErrMissingTenantPath, got %v", err)
	}
	if ErrorKindOf(err) != ErrorKindMissingRoute {
		t.Fatalf("expected missing route kind, got %s", ErrorKindOf(err))
	}
}

func TestQueryDetectsUnsupportedURLFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("unsupported URL format for path \"/prometheus/api/v1/query\""))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
	}))
	defer srv.Close()

	client := NewClient(domain.VMConnection{URL: srv.URL})
	_, err := client.Query(context.Background(), "vm_app_version", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingTenantPath) {
		t.Fatalf("expected ErrMissingTenantPath, got %v", err)
	}
	if ErrorKindOf(err) != ErrorKindMissingRoute {
		t.Fatalf("expected missing route kind, got %s", ErrorKindOf(err))
	}
}

func TestClassifyResponseError_QueryTimeout(t *testing.T) {
	err := classifyResponseError(http.StatusUnprocessableEntity, `{"status":"error","error":"timeout exceeded during query execution: 30.000 seconds"}`)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Kind != ErrorKindQueryTimeout {
		t.Fatalf("expected query timeout kind, got %s", apiErr.Kind)
	}
}

func TestClassifyResponseError_TooManySeries(t *testing.T) {
	err := classifyResponseError(http.StatusBadRequest, `the number of matching timeseries exceeds 10000000`)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Kind != ErrorKindTooManySeries {
		t.Fatalf("expected too many series kind, got %s", apiErr.Kind)
	}
}

func TestClassifyResponseError_Empty404KeepsStatusCode(t *testing.T) {
	err := classifyResponseError(http.StatusNotFound, "")
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected status code in error, got %q", err.Error())
	}
	if got := ErrorKindOf(err); got != ErrorKindMissingRoute {
		t.Fatalf("ErrorKindOf(empty 404) = %q, want %q", got, ErrorKindMissingRoute)
	}
}

func TestErrorKindOf_NotFoundAtStartIsMissingRoute(t *testing.T) {
	err := errors.New("not found: /api/v1/export")
	if got := ErrorKindOf(err); got != ErrorKindMissingRoute {
		t.Fatalf("ErrorKindOf(not found prefix) = %q, want %q", got, ErrorKindMissingRoute)
	}
}

func TestErrorKindOf_ContextDeadlineExceededIsQueryTimeout(t *testing.T) {
	if got := ErrorKindOf(context.DeadlineExceeded); got != ErrorKindQueryTimeout {
		t.Fatalf("ErrorKindOf(context deadline) = %q, want %q", got, ErrorKindQueryTimeout)
	}
}

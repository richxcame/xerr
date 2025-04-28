package xerr

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func init() {
	// Register a basic error for testing
	_ = LoadDefinitionsFromYAML("test_errors.yaml")
}

func TestHttpErrorBasics(t *testing.T) {
	err := BadRequest("bad_input", "Invalid input", "en", nil)

	if err.StatusCode() != http.StatusBadRequest {
		t.Error("unexpected status code")
	}
	if err.ErrorKey() != "bad_input" {
		t.Error("unexpected error key")
	}
	if !strings.Contains(err.Error(), "Invalid input") {
		t.Error("unexpected Error() message")
	}
}

func TestMetaFields(t *testing.T) {
	err := BadRequest("bad_input", "Invalid input", "en", nil)
	err.AddMeta("field", "email")

	value, ok := err.GetMeta("field")
	if !ok || value != "email" {
		t.Error("meta not stored or retrieved properly")
	}
}

func TestLocalizedMessageFallback(t *testing.T) {
	err := BadRequest("non_existing_key", "fallback message", "en", nil)
	if err.Localized() != "fallback message" {
		t.Error("expected fallback message")
	}
}

func TestValidationError(t *testing.T) {
	ve := ValidationError(map[string]string{"email": "required"})
	if ve.StatusCode() != http.StatusBadRequest {
		t.Error("expected 400")
	}
	fields, ok := ve.GetMeta("fields")
	if !ok || fields == nil {
		t.Error("validation fields missing")
	}
}

func TestErrorConversion(t *testing.T) {
	err := errors.New("some native error")
	httpErr := Internal(err, nil)

	unwrapped := errors.Unwrap(httpErr)
	if unwrapped.Error() != "some native error" {
		t.Error("unwrap not working correctly")
	}

	got, ok := AsHttpError(httpErr)
	if !ok || got.Key != "internal_error" {
		t.Error("AsHttpError failed")
	}
}

func TestIsHelper(t *testing.T) {
	err := NotFound("user_not_found", "User not found", "en")
	if !Is(err, "user_not_found") {
		t.Error("Is() helper failed")
	}
}

func TestResponderSuccess(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteAPIResponse(rr, map[string]string{"ok": "yes"}, nil, "en")

	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "success") {
		t.Error("success response failed")
	}
}

func TestResponderFailure(t *testing.T) {
	rr := httptest.NewRecorder()
	err := BadRequest("bad_input", "Invalid input", "en", nil)
	WriteAPIResponse(rr, nil, err, "en")

	if rr.Code != http.StatusBadRequest || !strings.Contains(rr.Body.String(), "bad_input") {
		t.Error("error response failed")
	}
}

func TestTraceContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Trace-ID", "abc")
	req.Header.Set("X-User-ID", "123")

	var capturedCtx context.Context
	handler := TraceContextMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if GetTraceID(capturedCtx) != "abc" || GetUserID(capturedCtx) != "123" {
		t.Error("trace/user context not injected properly")
	}
}

func TestPanicRecovery(t *testing.T) {
	handler := PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError || !strings.Contains(rr.Body.String(), "internal_error") {
		t.Error("panic not handled correctly")
	}
}

func TestYAMLRegistration(t *testing.T) {
	os.WriteFile("test_errors.yaml", []byte(`
- key: test_error
  code: 418
  default: I'm a teapot
  expose: true
  i18n:
    en: I'm a teapot
    fr: Je suis une théière
`), 0644)
	defer os.Remove("test_errors.yaml")

	err := LoadDefinitionsFromYAML("test_errors.yaml")
	if err != nil {
		t.Fatalf("failed to load definitions: %v", err)
	}

	def := Registry["test_error"]
	if def.Code != 418 || def.Default != "I'm a teapot" {
		t.Error("definition not loaded correctly")
	}
}

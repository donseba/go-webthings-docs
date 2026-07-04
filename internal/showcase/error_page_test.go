package showcase

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestErrorSectionRendersGoPartialErrorCard(t *testing.T) {
	handler := NewHandler(os.DirFS("../../deploy/website/showcase"))
	req := httptest.NewRequest(http.MethodGet, "/error/section", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body:\n%s", rec.Code, http.StatusOK, body)
	}
	if !strings.Contains(body, "Template render error") {
		t.Fatalf("expected go-partial error card, got:\n%s", body)
	}
	if strings.TrimSpace(body) == "error parsing templates: template: broken.gohtml:5: unexpected EOF" {
		t.Fatalf("got raw parse error instead of rendered error card")
	}
}

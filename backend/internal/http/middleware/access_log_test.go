package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestFormatAccessLogOmitsQueryString(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.test/health?access_token=secret-marker", nil)
	line := formatAccessLog(gin.LogFormatterParams{
		Request:    req,
		TimeStamp:  time.Unix(0, 0).UTC(),
		StatusCode: 200,
		Method:     "GET",
		ClientIP:   "192.0.2.1",
	})

	if strings.Contains(line, "secret-marker") || strings.Contains(line, "access_token") {
		t.Fatalf("access log leaked query string: %q", line)
	}
	if !strings.Contains(line, `"/health"`) {
		t.Fatalf("access log missing request path: %q", line)
	}
}

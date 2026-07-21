package handler

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMediaDimensionsAllowed(t *testing.T) {
	if !mediaDimensionsAllowed(8000, 8000) {
		t.Fatal("expected normal image dimensions to be allowed")
	}
	if mediaDimensionsAllowed(40000, 1) {
		t.Fatal("expected excessive dimension to be rejected")
	}
	if mediaDimensionsAllowed(20000, 20000) {
		t.Fatal("expected excessive pixel count to be rejected")
	}
}

func TestServeLocalMediaUsesDetectedMIMEForExtensionlessFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	path := filepath.Join(t.TempDir(), "content-hash-without-extension")
	content := []byte{0, 0, 0, 0, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f'}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write test media: %v", err)
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/i/hash", nil)

	serveLocalMedia(c, path, "image/avif")

	if got := recorder.Header().Get("Content-Type"); got != "image/avif" {
		t.Fatalf("expected image/avif content type, got %q", got)
	}
	if got := recorder.Body.Bytes(); string(got) != string(content) {
		t.Fatalf("unexpected response body: %v", got)
	}
}

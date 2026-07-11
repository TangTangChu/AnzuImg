package service

import "testing"

func TestProcessedMediaDimensionsAllowed(t *testing.T) {
	if !processedMediaDimensionsAllowed(8000, 8000) {
		t.Fatal("expected normal media dimensions to be allowed")
	}
	if processedMediaDimensionsAllowed(40000, 1) {
		t.Fatal("expected excessive dimension to be rejected")
	}
	if processedMediaDimensionsAllowed(20000, 20000) {
		t.Fatal("expected excessive pixel count to be rejected")
	}
}

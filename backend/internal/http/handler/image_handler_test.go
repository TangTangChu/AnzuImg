package handler

import "testing"

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

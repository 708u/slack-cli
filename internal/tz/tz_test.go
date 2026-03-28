package tz

import (
	"testing"
	"time"
)

func TestDefaultIsLocal(t *testing.T) {
	if Location() != time.Local {
		t.Errorf(
			"expected default Location to be time.Local, got %v",
			Location(),
		)
	}
}

func TestSetAndGet(t *testing.T) {
	original := Location()
	defer Set(original)

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load Asia/Tokyo: %v", err)
	}

	Set(tokyo)
	if Location() != tokyo {
		t.Errorf("expected Asia/Tokyo, got %v", Location())
	}

	Set(time.UTC)
	if Location() != time.UTC {
		t.Errorf("expected UTC, got %v", Location())
	}
}
